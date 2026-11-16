package readiness

import (
	"os/exec"
	"strings"
	"testing"
)

func TestReadinessProbeIsExecAndNotTCP(t *testing.T) {
	p := Probe()
	if p.Exec == nil || p.TCPSocket != nil {
		t.Fatalf("readiness probe must be an exec probe, got %+v", p.ProbeHandler)
	}
	script := strings.Join(p.Exec.Command, " ")
	for _, want := range []string{"nodetool info", "nodetool status", "nodetool tpstats", `"UN"`, "MutationStage"} {
		if !strings.Contains(script, want) {
			t.Errorf("readiness script must contain %q", want)
		}
	}
	if strings.Contains(script, "%d") {
		t.Errorf("pending mutations threshold was not substituted")
	}
}

func TestReadinessScriptSyntax(t *testing.T) {
	p := Probe()
	if out, err := exec.Command("bash", "-n", "-c", p.Exec.Command[2]).CombinedOutput(); err != nil {
		t.Fatalf("readiness script has invalid bash syntax: %v\n%s", err, out)
	}
}
