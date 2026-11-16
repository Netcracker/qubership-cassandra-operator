package readiness

import (
	"fmt"

	v13 "k8s.io/api/core/v1"
)

// maxPendingMutations is the largest MutationStage backlog (writes this replica has accepted
// but not yet applied) at which the pod still reports ready. A bigger backlog means the replica
// is lagging and must not be treated as a safe target for the next node drain.
const maxPendingMutations = 1000

// readinessScript reports ready (exit 0) only when the local node:
//  1. has gossip and the native transport (CQL) running,
//  2. is a member of the ring in state UN (Up/Normal, so not joining, leaving or moving),
//  3. has no large backlog of unapplied mutations.
//
// Only the LOCAL node is evaluated. Other nodes being down must not make this pod not ready,
// otherwise a single failure would make every pod not ready and take the whole cluster down.
// The probe is deliberately not used for liveness, so a failing check never restarts the pod.
const readinessScript = `
set -o pipefail
info=$(nodetool info 2>/dev/null) || exit 1
echo "$info" | grep -qiE '^Gossip active[[:space:]]*:[[:space:]]*true' || exit 1
echo "$info" | grep -qiE '^Native Transport active[[:space:]]*:[[:space:]]*true' || exit 1
id=$(echo "$info" | awk -F': *' '/^ID[[:space:]]*:/ {print $2; exit}')
[ -n "$id" ] || exit 1
state=$(nodetool status 2>/dev/null | awk -v id="$id" 'index($0, id) {print $1; exit}')
[ "$state" = "UN" ] || exit 1
pending=$(nodetool tpstats 2>/dev/null | awk '$1 == "MutationStage" {print $3; exit}')
[ -n "$pending" ] || exit 1
[ "$pending" -le %d ] || exit 1
exit 0
`

// Probe returns the Cassandra readiness probe. Failing it only removes the pod from
// service endpoints and blocks the PDB; it never restarts the container.
func Probe() *v13.Probe {
	return &v13.Probe{
		ProbeHandler: v13.ProbeHandler{
			Exec: &v13.ExecAction{
				Command: []string{"/bin/bash", "-c", fmt.Sprintf(readinessScript, maxPendingMutations)},
			},
		},
		InitialDelaySeconds: 100,
		// nodetool starts a JVM for each call, which is slow under small CPU limits.
		TimeoutSeconds:   30,
		PeriodSeconds:    20,
		SuccessThreshold: 1,
		FailureThreshold: 3,
	}
}
