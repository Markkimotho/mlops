package feature

import "testing"

func TestLookupReturnsPresentAndMissingEntities(t *testing.T) {
	store := NewMemoryStore()
	store.Put("user_features", "user_id=u1", map[string]any{"age": 42.0, "risk": 0.7})
	response, err := Lookup(store, Request{
		FeatureService: "user_features",
		Entities:       []map[string]any{{"user_id": "u1"}, {"user_id": "missing"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Results[0].Statuses["risk"] != "PRESENT" {
		t.Fatalf("unexpected first result: %#v", response.Results[0])
	}
	if response.Results[1].Statuses["entity"] != "NOT_FOUND" {
		t.Fatalf("unexpected missing result: %#v", response.Results[1])
	}
}
