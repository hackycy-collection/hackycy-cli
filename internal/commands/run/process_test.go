package run

import (
	"reflect"
	"testing"
)

func TestChildRequestUsesAnArgvPackageManagerInvocation(t *testing.T) {
	request := childRequest("/project", PackageManagerExternal, "check")
	want := ChildRequest{
		Executable: string(PackageManagerExternal),
		Arguments:  []string{"run", "check"},
		Directory:  "/project",
	}
	if !reflect.DeepEqual(request, want) {
		t.Fatalf("childRequest() = %#v, want %#v", request, want)
	}
}
