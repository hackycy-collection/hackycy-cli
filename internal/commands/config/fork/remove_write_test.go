package fork

import (
	"errors"
	"sync"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
)

func TestRemoveSelectedDelegatesToAppconfigAndKeepsMissingState(t *testing.T) {
	store := &recordingRemoveStore{removed: false}

	removed, err := RemoveSelected(store, "work")

	if err != nil || removed {
		t.Fatalf("RemoveSelected() = (%t, %v), want missing result", removed, err)
	}
	if got, want := store.names, []string{"work"}; !sameAddStrings(got, want) {
		t.Fatalf("remove names = %#v, want %#v", got, want)
	}
}

func TestRemoveSelectedPropagatesAppconfigFailure(t *testing.T) {
	failure := errors.New("publish configuration")

	removed, err := RemoveSelected(&recordingRemoveStore{removed: true, err: failure}, "work")

	if !removed || !errors.Is(err, failure) {
		t.Fatalf("RemoveSelected() = (%t, %v), want appconfig result with failure", removed, err)
	}
}

func TestRemoveSelectedPreservesConcurrentForkAndCMUpdates(t *testing.T) {
	store, _ := newAddStore(t)
	for _, input := range []AddInput{
		{Alias: "remove", Host: "remove.example", Type: "github", Scheme: "https", Token: "remove-token"},
		{Alias: "retain", Host: "retain.example", Type: "gitlab", Scheme: "http", Token: "retain-token"},
	} {
		if err := SaveAdd(store, input); err != nil {
			t.Fatalf("SaveAdd(%q) returned an error: %v", input.Alias, err)
		}
	}
	if err := store.AddCMProfile("existing", "https://existing.example", "model", "existing-key"); err != nil {
		t.Fatalf("AddCMProfile(existing) returned an error: %v", err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var group sync.WaitGroup
	group.Add(2)
	go func() {
		defer group.Done()
		<-start
		removed, err := RemoveSelected(store, "remove")
		if err == nil && !removed {
			results <- errors.New("selected Fork was not removed")
			return
		}
		results <- err
	}()
	go func() {
		defer group.Done()
		<-start
		results <- SaveAdd(store, AddInput{Alias: "concurrent", Host: "concurrent.example", Type: "github", Scheme: "https", Token: "concurrent-token"})
	}()
	close(start)
	group.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatalf("concurrent update returned an error: %v", err)
		}
	}

	instances, err := store.ListForkInstances()
	if err != nil {
		t.Fatalf("ListForkInstances() returned an error: %v", err)
	}
	if got, want := forkInstanceNames(instances), []string{"retain", "concurrent"}; !sameAddStrings(got, want) {
		t.Fatalf("Fork instances = %#v, want %#v", got, want)
	}
	profiles, err := store.ListCMProfiles()
	if err != nil || len(profiles.Profiles) != 1 || profiles.Profiles[0].Name != "existing" {
		t.Fatalf("ListCMProfiles() = (%#v, %v), want existing profile", profiles, err)
	}
}

type recordingRemoveStore struct {
	names   []string
	removed bool
	err     error
}

func (store *recordingRemoveStore) RemoveForkInstance(name string) (bool, error) {
	store.names = append(store.names, name)
	return store.removed, store.err
}

func forkInstanceNames(instances []appconfig.ForkInstance) []string {
	names := make([]string, len(instances))
	for index, instance := range instances {
		names[index] = instance.Name
	}
	return names
}
