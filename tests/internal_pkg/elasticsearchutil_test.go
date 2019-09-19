package tests

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/saelastic"
	"github.com/stretchr/testify/assert"
)

func TestSantizing(t *testing.T) {
	var unstructuredResult map[string]interface{}
	t.Run("Test Json Sanitizing for Connectivity", func(t *testing.T) {
		result1 := saelastic.Sanitize(connectivityDirty)
		err := json.Unmarshal([]byte(result1), &unstructuredResult)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, nil, err)
	})

	t.Run("Test Json Sanitizing for ProcEvent", func(t *testing.T) {
		result1 := saelastic.Sanitize(procEventDirty)
		err := json.Unmarshal([]byte(result1), &unstructuredResult)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, nil, err)
	})

	t.Run("Test Json Sanitizing for OVS", func(t *testing.T) {
		result1 := saelastic.Sanitize(ovsEventDirty)
		fmt.Println(result1)
		err := json.Unmarshal([]byte(result1), &unstructuredResult)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, nil, err)
	})
}
