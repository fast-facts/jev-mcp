package policy

const NoneOption = "none"

func AcceptPick(choice string, exists float64, ids map[string]struct{}) bool {
	if ExistsVerdict(exists) != ExistsAnsweredVerdict {
		return false
	}
	if choice == NoneOption {
		return false
	}
	_, ok := ids[choice]
	return ok
}
