package flow

import (
	"context"
	"fmt"
	"sync"
)

// Node interface represents a node in the flow.
type Node interface {
	run(context.Context, []Interceptor) error
	setNext(Node)
	getNext() Node
}

// base struct contains common fields for nodes.
type base struct {
	name string
}

// baseNode embeds base and contains the next node in the flow.
type baseNode struct {
	base
	next Node
}

// Set the next node.
func (n *baseNode) setNext(next Node) {
	n.next = next
}

// Get the next node.
func (n *baseNode) getNext() Node {
	return n.next
}

// doNode represents an action node that executes a function.
type doNode struct {
	baseNode
	fn func(context.Context) error
}

// Run executes the node's function and proceeds to the next node.
func (n *doNode) run(ctx context.Context, interceptors []Interceptor) error {
	for _, i := range interceptors {
		if err := i(ctx, n); err != nil {
			return err
		}
	}
	if err := n.fn(ctx); err != nil {
		return err
	}
	if n.next != nil {
		return n.next.run(ctx, interceptors)
	}
	return nil
}

// conditionalNode represents a node that branches based on a condition.
type conditionalNode struct {
	baseNode
	condition  func(context.Context) bool
	trueBranch Node
}

// Run evaluates the condition and executes the true branch if the condition is true.
func (n *conditionalNode) run(ctx context.Context, interceptors []Interceptor) error {
	for _, i := range interceptors {
		if err := i(ctx, n); err != nil {
			return err
		}
	}
	if n.condition(ctx) && n.trueBranch != nil {
		if err := n.trueBranch.run(ctx, interceptors); err != nil {
			return err
		}
	}
	// Proceed to the next node regardless of the condition result
	if n.next != nil {
		return n.next.run(ctx, interceptors)
	}
	return nil
}

// sequenceNode represents a sequence of nodes to be executed in order.
type sequenceNode struct {
	baseNode
	nodes []Node
}

// Run executes each node in the sequence.
func (n *sequenceNode) run(ctx context.Context, interceptors []Interceptor) error {
	for _, node := range n.nodes {
		if node != nil {
			if err := node.run(ctx, interceptors); err != nil {
				return err
			}
		}
	}
	if n.next != nil {
		return n.next.run(ctx, interceptors)
	}
	return nil
}

// Interceptor defines a function that can intercept node execution.
type Interceptor func(context.Context, Node) error

// Flow represents a sequence of nodes forming the DAG.
type Flow struct {
	base
	head             Node
	tail             Node
	flowInterceptors []Interceptor
	nodeInterceptors []Interceptor
}

// Ensure Flow implements Node by adding run, setNext, and getNext methods.
func (f *Flow) run(ctx context.Context, interceptors []Interceptor) error {
	if f.head == nil {
		return nil
	}
	// Run flow-level interceptors
	for _, i := range f.flowInterceptors {
		if err := i(ctx, nil); err != nil {
			return err
		}
	}
	// Start execution from the head node
	return f.head.run(ctx, f.nodeInterceptors)
}

func (f *Flow) setNext(next Node) {
	if f.tail != nil {
		f.tail.setNext(next)
		f.tail = next
	}
}

func (f *Flow) getNext() Node {
	if f.tail != nil {
		return f.tail.getNext()
	}
	return nil
}

// New creates a new flow with the given name.
func New(name string) *Flow {
	return &Flow{
		base: base{name: name},
	}
}

func (f *Flow) Name() string {
  return f.name
}

// Do adds a new action node to the flow.
func (f *Flow) Do(name string, fn func(context.Context) error) *Flow {
	node := &doNode{
		baseNode: baseNode{
			base: base{
				name: name,
			},
		},
		fn: fn,
	}
	f.appendNode(node)
	return f
}

// Then adds an existing node or flow to the current flow.
func (f *Flow) Then(node Node) *Flow {
	switch n := node.(type) {
	case *Flow:
		f.appendFlow(n)
	case Node:
		f.appendNode(n)
	default:
		panic(fmt.Sprintf("Then method accepts only Node or *Flow, got %T", node))
	}
	return f
}

func (f *Flow) appendFlow(flowNode *Flow) {
	if flowNode.head == nil {
		return
	}
	if f.head == nil {
		f.head = flowNode.head
		f.tail = flowNode.tail
	} else {
		f.tail.setNext(flowNode.head)
		f.tail = flowNode.tail
	}
}

// If adds a conditional node to the flow that executes the trueBranch if the condition is true.
func (f *Flow) If(name string, cond func(context.Context) bool, trueBranch Node) *Flow {
	condNode := &conditionalNode{
		baseNode: baseNode{
			base: base{
				name: name,
			},
		},
		condition:  cond,
		trueBranch: trueBranch,
	}
	f.appendNode(condNode)
	return f
}

// appendNode appends a node to the flow.
func (f *Flow) appendNode(node Node) {
	if f.head == nil {
		f.head = node
		f.tail = node
	} else {
		f.tail.setNext(node) // Ensures the tail is updated
		f.tail = node
	}
}

// InSequence creates a sequence node containing the provided nodes.
func InSequence(name string, nodes ...Node) Node {
	var filteredNodes []Node
	for _, node := range nodes {
		if node != nil {
			filteredNodes = append(filteredNodes, node)
		}
	}
	return &sequenceNode{
		baseNode: baseNode{
			base: base{
				name: name,
			},
		},
		nodes: filteredNodes,
	}
}

// Do creates a standalone action node.
func Do(name string, fn func(context.Context) error) Node {
	return &doNode{
		baseNode: baseNode{
			base: base{
				name: name,
			},
		},
		fn: fn,
	}
}

// Run starts executing the flow from the head node.
func (f *Flow) Run(ctx context.Context) error {
	if f.head == nil {
		return nil
	}
	// Run flow interceptors with the flow itself
	for _, i := range f.flowInterceptors {
		if err := i(ctx, nil); err != nil {
			return err
		}
	}
	// Start execution with the head node
	return f.head.run(ctx, f.nodeInterceptors)
}

// AddFlowInterceptor adds an interceptor that runs before the flow starts.
func (f *Flow) AddFlowInterceptor(i Interceptor) *Flow {
	f.flowInterceptors = append(f.flowInterceptors, i)
	return f
}

// AddNodeInterceptor adds an interceptor that runs before each node.
func (f *Flow) AddNodeInterceptor(i Interceptor) *Flow {
	f.nodeInterceptors = append(f.nodeInterceptors, i)
	return f
}

// parallelNode represents nodes that should be executed concurrently
type parallelNode struct {
	baseNode
	nodes []Node
}

// Run executes all nodes in parallel and waits for them to complete
func (n *parallelNode) run(ctx context.Context, interceptors []Interceptor) error {
	for _, i := range interceptors {
		if err := i(ctx, n); err != nil {
			return err
		}
	}

	errChan := make(chan error, len(n.nodes))
	var wg sync.WaitGroup
	wg.Add(len(n.nodes))

	for _, node := range n.nodes {
		go func(node Node) {
			defer wg.Done()
			if node != nil {
				if err := node.run(ctx, interceptors); err != nil {
					errChan <- err
				}
			}
		}(node)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	if n.next != nil {
		return n.next.run(ctx, interceptors)
	}
	return nil
}

// InParallel creates a parallel node containing the provided nodes
func InParallel(name string, nodes ...Node) Node {
	var filteredNodes []Node
	for _, node := range nodes {
		if node != nil {
			filteredNodes = append(filteredNodes, node)
		}
	}
	return &parallelNode{
		baseNode: baseNode{
			base: base{
				name: name,
			},
		},
		nodes: filteredNodes,
	}
}
