# End-to-End Testing Scripts

This directory contains scripts for comprehensive testing of the time tracker application.

## e2e-test.sh

A comprehensive end-to-end test script that verifies all major functionality of the time tracker application.

### Usage

```bash
./scripts/e2e-test.sh
```

### What It Tests

The script runs through a complete workflow testing all major features:

1. **Basic Task Management**
   - Creating tasks
   - Checking current task
   - Stopping tasks
   - Task switching

2. **Listing and Filtering**
   - List all tasks
   - Time-based filtering (e.g., `1h`, `2d`)
   - Text-based filtering
   - Combined filtering

3. **Data Export**
   - CSV export functionality
   - Data integrity in exports

4. **Resume Functionality**
   - Interactive task resumption
   - Proper task switching

5. **Summary Functionality**
   - Task summary generation
   - Interactive task selection

6. **Delete Functionality**
   - Task deletion
   - **Bug fix verification**: Ensures the delete command only removes the selected task and its entries (not all entries)

7. **Edge Cases**
   - No running tasks
   - Empty task names
   - Special characters

8. **Data Integrity**
   - Multiple task creation
   - Data consistency checks
   - Cross-functionality validation

9. **Performance Testing**
   - Response time validation
   - Multi-task handling

### Features

- **Colored Output**: Uses color coding for different message types (info, success, warning, error)
- **Automatic Cleanup**: Cleans up test artifacts on exit
- **Comprehensive Verification**: Validates both success cases and error conditions
- **Bug Regression Testing**: Specifically tests the delete bug fix
- **Performance Monitoring**: Measures and reports command execution times

### Test Database

The script uses a separate test database (`/tmp/tt_e2e_test.db`) to avoid interfering with production data.

### Exit Codes

- `0`: All tests passed successfully
- `1`: One or more tests failed

### When to Run

Run this script:
- **At the end of every development session** (as specified in the requirements)
- Before creating pull requests
- After making significant changes to the codebase
- Before releases
- When debugging reported issues

### Example Output

```
==============================================
  Time Tracker End-to-End Test Suite
==============================================

[INFO] Building test binary...
[SUCCESS] Binary built successfully

==========================================
  Test 1: Basic Task Management
==========================================
[SUCCESS] Task 1 creation - Expected text found
[SUCCESS] Current task display - Expected text found
...

==========================================
  Test Results Summary
==========================================
[SUCCESS] All end-to-end tests completed successfully!

Key functionality verified:
  ✓ Task creation and management
  ✓ Time tracking (start/stop/current)
  ✓ Task listing and filtering
  ✓ CSV data export
  ✓ Resume functionality
  ✓ Summary generation
  ✓ Delete functionality (bug fix verified)
  ✓ Edge case handling
  ✓ Data integrity
  ✓ Performance

[SUCCESS] Time tracker is ready for use!
```

### Troubleshooting

If the script fails:

1. **Check the error message** - The script provides detailed error information
2. **Review the actual vs expected output** - Failed assertions show what was expected
3. **Run individual commands manually** - Use the same `TT_DB_DIR=/tmp TT_DB_FILENAME=tt_e2e_test.db` prefix
4. **Check for recent code changes** - The failure might indicate a regression

### Maintenance

When adding new features to the time tracker:

1. Add corresponding tests to this script
2. Update the test summary section
3. Ensure new functionality is covered by appropriate assertions
4. Test the script after making changes

The script is designed to be easily extensible - just add new test sections following the existing pattern.