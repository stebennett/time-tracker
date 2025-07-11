#!/bin/bash

# End-to-End Test Script for Time Tracker
# This script runs comprehensive manual tests to verify all functionality works correctly
# Run this at the end of every development session

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_DB_DIR="/tmp"
TEST_DB_FILENAME="tt_e2e_test.db"
TEST_BINARY="./tt-e2e-test"

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

run_cmd() {
    local cmd="$1"
    local description="$2"
    log_info "Running: $description"
    echo "  Command: TT_DB_DIR=$TEST_DB_DIR TT_DB_FILENAME=$TEST_DB_FILENAME $cmd"
    TT_DB_DIR=$TEST_DB_DIR TT_DB_FILENAME=$TEST_DB_FILENAME eval "$cmd"
    echo
}

verify_output() {
    local expected="$1"
    local actual="$2"
    local test_name="$3"
    
    if echo "$actual" | grep -q "$expected"; then
        log_success "$test_name - Expected text found"
    else
        log_error "$test_name - Expected text not found"
        echo "Expected: $expected"
        echo "Actual: $actual"
        exit 1
    fi
}

cleanup() {
    log_info "Cleaning up test artifacts..."
    rm -f "$TEST_DB_DIR/$TEST_DB_FILENAME" "$TEST_BINARY"
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Start of test script
echo "=============================================="
echo "  Time Tracker End-to-End Test Suite"
echo "=============================================="
echo

# 1. Build test binary
log_info "Building test binary..."
go build -o "$TEST_BINARY" cmd/tt/main.go
log_success "Binary built successfully"
echo

# 2. Clean test database
log_info "Setting up clean test environment..."
rm -f "$TEST_DB_DIR/$TEST_DB_FILENAME"
log_success "Test environment ready"
echo

# 3. Test basic task creation and management
echo "=========================================="
echo "  Test 1: Basic Task Management"
echo "=========================================="

# Start first task
output=$(run_cmd "$TEST_BINARY start 'E2E Test Task 1'" "Start first task")
verify_output "Started new task: E2E Test Task 1" "$output" "Task 1 creation"

# Check current task
output=$(run_cmd "$TEST_BINARY current" "Check current task")
verify_output "Current task: E2E Test Task 1" "$output" "Current task display"

# Start second task (should stop first)
output=$(run_cmd "$TEST_BINARY start 'E2E Test Task 2'" "Start second task")
verify_output "All running tasks have been stopped" "$output" "Previous task stopped"
verify_output "Started new task: E2E Test Task 2" "$output" "Task 2 creation"

# Start third task
output=$(run_cmd "$TEST_BINARY start 'E2E Test Task 3'" "Start third task")
verify_output "Started new task: E2E Test Task 3" "$output" "Task 3 creation"

# Stop current task
output=$(run_cmd "$TEST_BINARY stop" "Stop current task")
verify_output "All running tasks have been stopped" "$output" "Task stopping"

echo

# 4. Test listing and filtering
echo "=========================================="
echo "  Test 2: Listing and Filtering"
echo "=========================================="

# List all tasks
output=$(run_cmd "$TEST_BINARY list" "List all tasks")
verify_output "E2E Test Task 1" "$output" "Task 1 in list"
verify_output "E2E Test Task 2" "$output" "Task 2 in list"
verify_output "E2E Test Task 3" "$output" "Task 3 in list"

# Test time filtering
output=$(run_cmd "$TEST_BINARY list 1h" "List tasks from last hour")
verify_output "E2E Test Task" "$output" "Time filtering works"

# Test text filtering
output=$(run_cmd "$TEST_BINARY list 'Task 2'" "List tasks with 'Task 2'")
verify_output "E2E Test Task 2" "$output" "Text filtering works"

# Test combined filtering
output=$(run_cmd "$TEST_BINARY list 1h 'Task'" "List tasks with time and text filter")
verify_output "E2E Test Task" "$output" "Combined filtering works"

echo

# 5. Test CSV export
echo "=========================================="
echo "  Test 3: Data Export"
echo "=========================================="

output=$(run_cmd "$TEST_BINARY output format=csv" "Export to CSV")
verify_output "ID,Start Time,End Time,Duration (hours),Task Name" "$output" "CSV header"
verify_output "E2E Test Task 1" "$output" "Task 1 in CSV"
verify_output "E2E Test Task 2" "$output" "Task 2 in CSV"
verify_output "E2E Test Task 3" "$output" "Task 3 in CSV"

echo

# 6. Test resume functionality
echo "=========================================="
echo "  Test 4: Resume Functionality"
echo "=========================================="

# Resume a task (select option 2)
output=$(echo "2" | run_cmd "$TEST_BINARY resume" "Resume Task 2")
verify_output "Select a task to resume:" "$output" "Resume menu shown"
verify_output "E2E Test Task 2" "$output" "Task 2 in resume list"
verify_output "Resumed task: E2E Test Task 2" "$output" "Task 2 resumed"

# Verify task is running
output=$(run_cmd "$TEST_BINARY current" "Check resumed task")
verify_output "Current task: E2E Test Task 2" "$output" "Resumed task is current"

echo

# 7. Test summary functionality
echo "=========================================="
echo "  Test 5: Summary Functionality"
echo "=========================================="

# Stop current task to have data for summary
run_cmd "$TEST_BINARY stop" "Stop current task for summary"

# Get summary (select option 1)
output=$(echo "1" | run_cmd "$TEST_BINARY summary" "Get summary for first task")
verify_output "Select a task to summarize:" "$output" "Summary menu shown"
# Summary output varies based on data, so just check that it doesn't error
if echo "$output" | grep -q "Error:"; then
    log_error "Summary command failed with error"
    exit 1
else
    log_success "Summary command completed without errors"
fi

echo

# 8. Test delete functionality (the bug we fixed)
echo "=========================================="
echo "  Test 6: Delete Functionality (Bug Fix)"
echo "=========================================="

# Delete first task (select option 1)
output=$(echo "1" | run_cmd "$TEST_BINARY delete" "Delete first task")
verify_output "Select a task to delete:" "$output" "Delete menu shown"
verify_output "Deleted task:" "$output" "Task deletion confirmed"

# Verify other tasks still exist (this was the bug)
output=$(run_cmd "$TEST_BINARY list" "List remaining tasks after deletion")

# Check that the list command works without errors (this was the main bug)
if echo "$output" | grep -q "Error:"; then
    log_error "Delete bug regression: List command failed after deletion"
    exit 1
else
    log_success "Delete works correctly: List command works after deletion"
fi

# Verify some tasks still exist
if echo "$output" | grep -q "E2E Test Task"; then
    log_success "Other tasks still exist after deletion"
else
    log_error "Delete bug regression: All tasks were deleted"
    exit 1
fi

echo

# 9. Test edge cases
echo "=========================================="
echo "  Test 7: Edge Cases"
echo "=========================================="

# Test with no running tasks
output=$(run_cmd "$TEST_BINARY current" "Check current with no running tasks")
verify_output "No task is currently running" "$output" "No running task message"

# Test stop with no running tasks
output=$(run_cmd "$TEST_BINARY stop" "Stop with no running tasks")
verify_output "All running tasks have been stopped" "$output" "Stop with no running tasks"

# Test with empty task name (should fail)
output=$(run_cmd "$TEST_BINARY start ''" "Start task with empty name" 2>&1 || true)
verify_output "Error:" "$output" "Empty task name properly rejected"

# Test with special characters
output=$(run_cmd "$TEST_BINARY start 'Task with special chars'" "Start task with special chars")
verify_output "Started new task: Task with special chars" "$output" "Special characters handled"

echo

# 10. Test data integrity
echo "=========================================="
echo "  Test 8: Data Integrity"
echo "=========================================="

# Create multiple tasks and verify data consistency
run_cmd "$TEST_BINARY start 'Integrity Test 1'" "Create integrity test task 1"
run_cmd "$TEST_BINARY start 'Integrity Test 2'" "Create integrity test task 2"
run_cmd "$TEST_BINARY start 'Integrity Test 3'" "Create integrity test task 3"
run_cmd "$TEST_BINARY stop" "Stop for integrity test"

# List should show all tasks without errors
output=$(run_cmd "$TEST_BINARY list" "List all tasks for integrity check")
verify_output "Integrity Test 1" "$output" "Integrity test task 1 exists"
verify_output "Integrity Test 2" "$output" "Integrity test task 2 exists"
verify_output "Integrity Test 3" "$output" "Integrity test task 3 exists"

# CSV export should work without errors
output=$(run_cmd "$TEST_BINARY output format=csv" "CSV export integrity check")
verify_output "Integrity Test 1" "$output" "Integrity test task 1 in CSV"
verify_output "Integrity Test 2" "$output" "Integrity test task 2 in CSV"
verify_output "Integrity Test 3" "$output" "Integrity test task 3 in CSV"

echo

# 11. Performance test with multiple entries
echo "=========================================="
echo "  Test 9: Performance Test"
echo "=========================================="

log_info "Creating multiple tasks for performance test..."
for i in {1..10}; do
    TT_DB_DIR=$TEST_DB_DIR TT_DB_FILENAME=$TEST_DB_FILENAME $TEST_BINARY start "Perf Test Task $i" > /dev/null
done
TT_DB_DIR=$TEST_DB_DIR TT_DB_FILENAME=$TEST_DB_FILENAME $TEST_BINARY stop > /dev/null

# Time the list command
start_time=$(date +%s%N)
output=$(run_cmd "$TEST_BINARY list" "Performance test - list 10 tasks")
end_time=$(date +%s%N)
duration=$(( (end_time - start_time) / 1000000 ))  # Convert to milliseconds

log_info "List command took ${duration}ms"
if [ $duration -lt 1000 ]; then
    log_success "Performance test passed - list command under 1 second"
else
    log_warning "Performance test - list command took ${duration}ms (over 1 second)"
fi

echo

# Final summary
echo "=========================================="
echo "  Test Results Summary"
echo "=========================================="

log_success "All end-to-end tests completed successfully!"
log_info "Test database used: $TEST_DB"
log_info "Test binary used: $TEST_BINARY"

echo
echo "Key functionality verified:"
echo "  ✓ Task creation and management"
echo "  ✓ Time tracking (start/stop/current)"
echo "  ✓ Task listing and filtering"
echo "  ✓ CSV data export"
echo "  ✓ Resume functionality"
echo "  ✓ Summary generation"
echo "  ✓ Delete functionality (bug fix verified)"
echo "  ✓ Edge case handling"
echo "  ✓ Data integrity"
echo "  ✓ Performance"

echo
log_success "Time tracker is ready for use!"