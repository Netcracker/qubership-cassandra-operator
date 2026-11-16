*** Variables ***
${PDB_NAME}                      cassandra-cluster
${POD_SELECTOR}                  service=cassandra-cluster
${READY_RETRY_TIME}              ${WAIT_TIMEOUT}s
${READY_RETRY_INTERVAL}          10s

*** Settings ***
Resource  ../shared/keywords.robot
Library  OperatingSystem
Library  String
Suite Setup  Prepare Shared

*** Keywords ***
Kubectl
    [Documentation]  Runs kubectl in the Cassandra namespace and returns stdout. Fails on non-zero exit code.
    [Arguments]    ${args}
    ${rc}  ${out}=  Run And Return Rc And Output  kubectl -n ${CASSANDRA_NAMESPACE} ${args}
    Should Be Equal As Integers  ${rc}  0  kubectl ${args} failed: ${out}
    RETURN    ${out}

Get Disruptions Allowed
    ${out}=  Kubectl  get pdb ${PDB_NAME} -o jsonpath={.status.disruptionsAllowed}
    RETURN    ${out}

Get Pod Uid
    [Arguments]    ${pod}
    ${uid}=  Kubectl  get pod ${pod} -o jsonpath={.metadata.uid}
    RETURN    ${uid}

Evict Pod
    [Arguments]    ${pod}
    ${body}=  Set Variable  {"apiVersion":"policy/v1","kind":"Eviction","metadata":{"name":"${pod}","namespace":"${CASSANDRA_NAMESPACE}"}}
    ${rc}  ${out}=  Run And Return Rc And Output  echo '${body}' | kubectl create --raw /api/v1/namespaces/${CASSANDRA_NAMESPACE}/pods/${pod}/eviction -f -
    RETURN    ${rc}    ${out}

Pod Is Recreated And Ready
    [Arguments]    ${pod}    ${old_uid}
    ${uid}=  Get Pod Uid  ${pod}
    Should Not Be Equal  ${uid}  ${old_uid}  Pod ${pod} was not recreated yet
    ${ready}=  Kubectl  get pod ${pod} -o jsonpath={.status.conditions[?(@.type=="Ready")].status}
    Should Be Equal As Strings  ${ready}  True  Pod ${pod} is not ready yet

All Ring Nodes Are Up Normal
    [Arguments]    ${pod}
    ${status}=  Execute Command In Pod  ${pod}  ${CASSANDRA_NAMESPACE}  nodetool status
    Should Not Match Regexp  ${status}[0]  (?m)^(DN|UJ|UL|UM)\\s  nodetool status reports nodes that are not UN: ${status}[0]

*** Test Cases ***
Test Node Drain Waits For Pod To Rejoin Ring
    [Documentation]  Evicts one Cassandra pod the way a node drain does and checks that the PDB stops further
    ...              disruptions until the pod is recreated, ready (UN in the ring) and the budget is restored.
    [Tags]  ha  cassandra  drain
    Skip If  ${TEST_KEYSPACES_REPLICATION_FACTOR} < 3  Replicas < 3, PDB is not created, not possible to check case!
    ${allowed_before}=  Get Disruptions Allowed
    Should Be True  ${allowed_before} >= 1  PDB must allow at least one disruption on a healthy cluster
    ${pods}=  Kubectl  get pods -l ${POD_SELECTOR} -o jsonpath={.items[*].metadata.name}
    @{names}=  Split String  ${pods}
    ${victim}=  Set Variable  ${names}[0]
    ${old_uid}=  Get Pod Uid  ${victim}

    ${rc}  ${out}=  Evict Pod  ${victim}
    Should Be Equal As Integers  ${rc}  0  First eviction must be allowed: ${out}

    # While the evicted pod is gone or not ready, the budget must shrink and the next drain must be held back.
    Wait Until Keyword Succeeds  2min  5s  Disruptions Are Reduced  ${allowed_before}

    Wait Until Keyword Succeeds  ${READY_RETRY_TIME}  ${READY_RETRY_INTERVAL}  Pod Is Recreated And Ready  ${victim}  ${old_uid}
    Wait Until Keyword Succeeds  ${READY_RETRY_TIME}  ${READY_RETRY_INTERVAL}  All Ring Nodes Are Up Normal  ${victim}
    Wait Until Keyword Succeeds  2min  5s  Disruptions Are Restored  ${allowed_before}

*** Keywords ***
Disruptions Are Reduced
    [Arguments]    ${allowed_before}
    ${now}=  Get Disruptions Allowed
    Should Be True  ${now} < ${allowed_before}  PDB still allows ${now} disruptions, expected fewer than ${allowed_before}

Disruptions Are Restored
    [Arguments]    ${allowed_before}
    ${now}=  Get Disruptions Allowed
    Should Be True  ${now} >= ${allowed_before}  PDB allows ${now} disruptions, expected at least ${allowed_before}
