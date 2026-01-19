#!/bin/bash
set -e

echo "Testing llm-todo with imported tasks..."

TODO=/Users/daniel.nitsch/ais/projects/llm-todo/bin/todo
TODODIR=/Users/daniel.nitsch/ais/projects/llm-todo/todo

# Work in llm-todo directory (session auto-detected from dir name)
cd /Users/daniel.nitsch/ais/projects/llm-todo

echo "1. Import llm-todo's own p0 tasks..."
$TODO import $TODODIR/todo.p0.yaml

echo ""
echo "2. Get p0 tasks (minimal - saves tokens)..."
$TODO get p0 | head -12

echo ""
echo "3. Show task 4 (conditional formatter)..."
$TODO show 4 | head -15

echo ""
echo "4. Complete task and check for unblocked..."
$TODO done 1

echo ""
echo "5. Session status..."
$TODO status

echo ""
echo "✓ Import test complete!"
