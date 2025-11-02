package domain

type ExecutionPolicy int

const (
	PolicyOptimized ExecutionPolicy = iota
	PolicyExhaustive
)

var policyToString = map[ExecutionPolicy]string{
	PolicyOptimized: "Optimized",
	PolicyExhaustive: "Exhaustive",
}

const defaultPolicy = PolicyExhaustive

var stringToPolicy map[string]ExecutionPolicy

func init() {
	stringToPolicy = make(map[string]ExecutionPolicy)
	for policy, str := range policyToString {
		stringToPolicy[str] = policy
	}
}

func (e ExecutionPolicy) String() string {
	if str, exists := policyToString[e]; exists {
		return str
	}
	return "Unknown"
}

func ParseExecutionPolicy(s string) (ExecutionPolicy, bool) {
	policy, exists := stringToPolicy[s]
	return policy, exists
}

func ParseExecutionPolicyWithDefault(s string) ExecutionPolicy {
	if policy, exists := stringToPolicy[s]; exists {
		return policy
	}
	return defaultPolicy
}