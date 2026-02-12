package validate

import (
	"fmt"
	"sort"
	"strings"
)

var allowedValues = map[string][]string{
	"transaction.history.version":       {"v1", "v2"},
	"transaction.history.sort":          {"ASCENDING", "DESCENDING"},
	"transaction.history.productType":   {"AUTO_RENEWABLE", "NON_RENEWABLE", "CONSUMABLE", "NON_CONSUMABLE"},
	"transaction.history.ownershipType": {"FAMILY_SHARED", "PURCHASED"},
	"subscription.status":               {"1", "2", "3", "4", "5"},
}

func Allowed(flag string) []string {
	values := allowedValues[flag]
	copyValues := make([]string, 0, len(values))
	copyValues = append(copyValues, values...)
	sort.Strings(copyValues)
	return copyValues
}

func NormalizeOne(flag string, value string) (string, error) {
	values, ok := allowedValues[flag]
	if !ok || value == "" {
		return value, nil
	}
	for _, allowed := range values {
		if strings.EqualFold(allowed, value) {
			return allowed, nil
		}
	}
	return "", fmt.Errorf("invalid %s: %s (allowed: %s)", flag, value, strings.Join(Allowed(flag), ", "))
}

func NormalizeMany(flag string, values []string) ([]string, error) {
	if len(values) == 0 {
		return values, nil
	}
	output := make([]string, 0, len(values))
	for _, value := range values {
		normalized, err := NormalizeOne(flag, value)
		if err != nil {
			return nil, err
		}
		output = append(output, normalized)
	}
	return output, nil
}
