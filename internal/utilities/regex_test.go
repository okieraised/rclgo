package utilities

import "testing"

func TestSplitMsgDefaultArrayValues(t *testing.T) {
	tests := []struct {
		name      string
		ros2type  string
		input     string
		wantSlice []string
	}{
		{
			name:      "Empty brackets -> nil slice",
			ros2type:  "int32",
			input:     "[]",
			wantSlice: nil,
		},
		{
			name:      "Non-string numbers with spaces",
			ros2type:  "int32",
			input:     "[1, 2, 3]",
			wantSlice: []string{"1", "2", "3"},
		},
		{
			name:      "String types: single-quoted items get normalized and re-quoted",
			ros2type:  "string",
			input:     "['foo', 'bar baz']",
			wantSlice: []string{`"foo"`, `"bar baz"`},
		},
		{
			name:      "String types: mixed quotes and embedded quotes",
			ros2type:  "string",
			input:     `["he said \"hi\"", 'it\'s ok']`,
			wantSlice: []string{`"he said \"hi\""`, `"it's ok"`},
		},
		{
			name:      "Whitespace and newlines tolerated",
			ros2type:  "uint8",
			input:     "[ 10 ,\n 20,\t30 ]",
			wantSlice: []string{"10", "20", "30"},
		},
		{
			name:      "Empty string element becomes quoted empty",
			ros2type:  "string",
			input:     "['']",
			wantSlice: []string{`""`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SplitMsgDefaultArrayValues(tc.ros2type, tc.input)

			// compare nil vs empty explicitly
			if tc.wantSlice == nil && got != nil {
				t.Fatalf("want nil slice, got %v", got)
			}
			if tc.wantSlice != nil {
				if got == nil {
					t.Fatalf("want %v, got nil", tc.wantSlice)
				}
				if len(got) != len(tc.wantSlice) {
					t.Fatalf("length mismatch: want %d, got %d (%v)", len(tc.wantSlice), len(got), got)
				}
				for i := range got {
					if got[i] != tc.wantSlice[i] {
						t.Fatalf("at %d: want %q, got %q", i, tc.wantSlice[i], got[i])
					}
				}
			}
		})
	}
}

func TestSrvNameFromSrvMsgName_ROS2(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"example_interfaces/srv/AddTwoInts_Request", "example_interfaces/srv/AddTwoInts"},
		{"example_interfaces/srv/AddTwoInts_Response", "example_interfaces/srv/AddTwoInts"},
		{"std_srvs/srv/Trigger_Request", "std_srvs/srv/Trigger"},
		{"rcl_interfaces/srv/SetParameters_Response", "rcl_interfaces/srv/SetParameters"},
		{"example_interfaces/srv/Empty", "example_interfaces/srv/Empty"}, // unchanged (no suffix)
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := SrvNameFromSrvMsgName(tc.in); got != tc.want {
				t.Fatalf("SrvNameFromSrvMsgName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestActionNameFromActionMsgName_ROS2(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"action_tutorials_interfaces/action/Fibonacci_Goal", "action_tutorials_interfaces/action/Fibonacci"},
		{"action_tutorials_interfaces/action/Fibonacci_Result", "action_tutorials_interfaces/action/Fibonacci"},
		{"action_tutorials_interfaces/action/Fibonacci_Feedback", "action_tutorials_interfaces/action/Fibonacci"},
		{"action_tutorials_interfaces/action/Fibonacci_SendGoal_Request", "action_tutorials_interfaces/action/Fibonacci"},
		{"action_tutorials_interfaces/action/Fibonacci_SendGoal_Response", "action_tutorials_interfaces/action/Fibonacci"},
		{"action_tutorials_interfaces/action/Fibonacci_GetResult_Request", "action_tutorials_interfaces/action/Fibonacci"},
		{"action_tutorials_interfaces/action/Fibonacci_GetResult_Response", "action_tutorials_interfaces/action/Fibonacci"},
		{"action_tutorials_interfaces/action/Fibonacci_FeedbackMessage", "action_tutorials_interfaces/action/Fibonacci"},
		{"nav2_msgs/action/NavigateToPose_Goal", "nav2_msgs/action/NavigateToPose"},
		{"control_msgs/action/GripperCommand_Feedback", "control_msgs/action/GripperCommand"},
		{"action_tutorials_interfaces/action/Fibonacci_SomethingElse", "action_tutorials_interfaces/action/Fibonacci_SomethingElse"}, // unchanged
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := ActionNameFromActionMsgName(tc.in); got != tc.want {
				t.Fatalf("ActionNameFromActionMsgName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestActionNameFromActionSrvName_ROS2(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"action_tutorials_interfaces/action/Fibonacci_SendGoal", "action_tutorials_interfaces/action/Fibonacci"},
		{"action_tutorials_interfaces/action/Fibonacci_GetResult", "action_tutorials_interfaces/action/Fibonacci"},
		{"nav2_msgs/action/NavigateToPose_SendGoal", "nav2_msgs/action/NavigateToPose"},
		{"nav2_msgs/action/NavigateToPose_GetResult", "nav2_msgs/action/NavigateToPose"},
		{"control_msgs/action/GripperCommand_SendGoalX", "control_msgs/action/GripperCommand_SendGoalX"}, // unchanged
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := ActionNameFromActionSrvName(tc.in); got != tc.want {
				t.Fatalf("ActionNameFromActionSrvName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestCReturnCodeNameToGo_ROS2(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		// rcl return codes
		{"RCL_RET_OK", "Ok"},
		{"RCL_RET_INVALID_ARGUMENT", "InvalidArgument"},
		{"RCL_RET_BAD_ALLOC", "BadAlloc"},
		{"RCL_RET_TIMEOUT", "Timeout"},
		// rmw return codes
		{"RMW_RET_OK", "RmwOk"},
		{"RMW_RET_ERROR", "RmwError"},
		{"RMW_RET_TIMEOUT", "RmwTimeout"},
		// unchanged prefixes (not matching)
		{"RCUTILS_RET_OK", "RcutilsRetOk"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := CReturnCodeNameToGo(tc.in); got != tc.want {
				t.Fatalf("CReturnCodeNameToGo(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
