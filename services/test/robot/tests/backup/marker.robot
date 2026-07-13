*** Variables ***
${BACKUP_HOST}                                    %{BACKUP_HOST}

*** Settings ***
Library  String
Library  Collections
Library  RequestsLibrary
Library  OperatingSystem
Resource  ../shared/keywords.robot
Suite Setup  Preparation

*** Keywords ***
Preparation
    Prepare Shared
    ${verify}=    Get Environment Variable    name=TLS_ROOTCERT    default=False
    Set Suite Variable  ${verify}
    ${backup_tls}=    Get Environment Variable    name=TLS_ENABLED    default=False
    ${port}=    Get Environment Variable    name=PORT    default=8080
    Set Suite Variable  ${port}

    ${PROTOCOL} =    Set Variable If    '${backup_tls}' == 'true'
    ...  https
    ...  http
    Set Suite Variable  ${PROTOCOL}

    Create Session    markersession    ${PROTOCOL}://${BACKUP_DAEMON_API_CREDENTIALS_USERNAME}:${BACKUP_DAEMON_API_CREDENTIALS_PASSWORD}@${BACKUP_HOST}:${port}   verify=${verify}

*** Test Cases ***
Test Set Marker
    [Tags]  backup  cassandra
    ${marker_value}=    Set Variable    my-backup/2026-07-07T17:15:00Z
    ${body}=    Set Variable    {"marker": "${marker_value}"}
    ${resp}=    POST On Session    markersession    /marker    data=${body}    headers=${headers}    expected_status=200
    Log    ${resp.content}
    Should Be Equal As Strings    ${resp.status_code}    200

Test Get Marker Returns Single Record
    [Tags]  backup  cassandra
    ${resp}=    GET On Session    markersession    /marker    expected_status=200
    Log    ${resp.content}
    Should Be Equal As Strings    ${resp.status_code}    200
    Should Contain    ${resp.content}    marker
    Should Not Contain    ${resp.content}    Debug:

Test Set Marker Replaces Existing
    [Tags]  backup  cassandra
    ${first_value}=    Set Variable    my-backup/2026-01-01T00:00:00Z
    ${body}=    Set Variable    {"marker": "${first_value}"}
    POST On Session    markersession    /marker    data=${body}    headers=${headers}    expected_status=200

    ${second_value}=    Set Variable    my-backup/2026-07-07T17:15:00Z
    ${body}=    Set Variable    {"marker": "${second_value}"}
    POST On Session    markersession    /marker    data=${body}    headers=${headers}    expected_status=200

    ${resp}=    GET On Session    markersession    /marker    expected_status=200
    Log    ${resp.content}
    Should Contain    ${resp.content}    ${second_value}
    Should Not Contain    ${resp.content}    ${first_value}

Test Get Marker Response Contains Only Marker Value
    [Tags]  backup  cassandra
    ${marker_value}=    Set Variable    my-backup/2026-07-07T17:15:00Z
    ${body}=    Set Variable    {"marker": "${marker_value}"}
    POST On Session    markersession    /marker    data=${body}    headers=${headers}    expected_status=200

    ${resp}=    GET On Session    markersession    /marker    expected_status=200
    Log    ${resp.content}
    Should Contain    ${resp.content}    ${marker_value}
    Should Not Contain    ${resp.content}    Debug:
    Should Not Contain    ${resp.content}    host=

Test Wrong Marker Credentials
    [Tags]  backup  cassandra
    ${wronguser}=    Generate Random String    10
    ${wrongpass}=    Generate Random String    10
    Create Session    wrongmarkersess    ${PROTOCOL}://${wronguser}:${wrongpass}@${BACKUP_HOST}:${port}    verify=${verify}
    ${resp}=    GET On Session    wrongmarkersess    /marker    expected_status=401
    Should Be Equal As Strings    ${resp.status_code}    401
