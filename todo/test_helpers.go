package todo

// Test helpers using functional options pattern

// withClock sets the clock for the todo list (for testing)
func withClock(clock Clock) Option {
	return func(tl *TodoList) {
		tl.clock = clock
	}
}

// withOrphanTodo adds a todo with a non-existent parent ID
func withOrphanTodo(id, title, fakeParentID string) Option {
	return func(tl *TodoList) {
		todo := &Todo{
			ID:        id,
			Title:     title,
			ParentID:  fakeParentID,
			Children:  []*Todo{},
			CreatedAt: tl.clock.Now(),
			UpdatedAt: tl.clock.Now(),
		}
		tl.todos[id] = todo
	}
}

// withRootWithParentID adds a todo to roots slice that has a non-empty ParentID
func withRootWithParentID(id, title, fakeParentID string) Option {
	return func(tl *TodoList) {
		todo := &Todo{
			ID:        id,
			Title:     title,
			ParentID:  fakeParentID,
			Children:  []*Todo{},
			CreatedAt: tl.clock.Now(),
			UpdatedAt: tl.clock.Now(),
		}
		tl.todos[id] = todo
		tl.roots = append(tl.roots, todo)
	}
}

// withChildNotInParentChildren adds a child to the map but not to parent's Children slice
func withChildNotInParentChildren(childID, childTitle, parentID string) Option {
	return func(tl *TodoList) {
		child := &Todo{
			ID:        childID,
			Title:     childTitle,
			ParentID:  parentID,
			Children:  []*Todo{},
			CreatedAt: tl.clock.Now(),
			UpdatedAt: tl.clock.Now(),
		}
		tl.todos[childID] = child
		// Intentionally NOT adding to parent.Children
	}
}

// withGhostChild adds a child to parent's Children slice but not to the map
func withGhostChild(parentID, ghostID, ghostTitle string) Option {
	return func(tl *TodoList) {
		parent := tl.todos[parentID]
		if parent == nil {
			return
		}
		ghost := &Todo{
			ID:        ghostID,
			Title:     ghostTitle,
			ParentID:  parentID,
			Children:  []*Todo{},
			CreatedAt: tl.clock.Now(),
			UpdatedAt: tl.clock.Now(),
		}
		parent.Children = append(parent.Children, ghost)
		// Intentionally NOT adding to tl.todos
	}
}

// withSelfParent adds a todo that references itself as parent
func withSelfParent(id, title string) Option {
	return func(tl *TodoList) {
		todo := &Todo{
			ID:        id,
			Title:     title,
			ParentID:  id, // Self-reference
			Children:  []*Todo{},
			CreatedAt: tl.clock.Now(),
			UpdatedAt: tl.clock.Now(),
		}
		tl.todos[id] = todo
	}
}

// withCycle creates a 3-node cycle: A -> B -> C -> A
// Requires nodes to exist first via Add()
func withCycle(nodeAID, nodeBID, nodeCID string) Option {
	return func(tl *TodoList) {
		nodeA := tl.todos[nodeAID]
		nodeB := tl.todos[nodeBID]
		nodeC := tl.todos[nodeCID]
		if nodeA == nil || nodeB == nil || nodeC == nil {
			return
		}

		// Remove A from roots (it's becoming a child)
		newRoots := []*Todo{}
		for _, root := range tl.roots {
			if root.ID != nodeAID {
				newRoots = append(newRoots, root)
			}
		}
		tl.roots = newRoots

		// Create cycle: C -> A
		nodeA.ParentID = nodeCID
		nodeC.Children = append(nodeC.Children, nodeA)
	}
}
