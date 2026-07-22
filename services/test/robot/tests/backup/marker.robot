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
    ${resp}=    POST On Session    markersession    /api/v1/data-validation/marker    data=${body}    headers=${headers}    expected_status=201
    Log    ${resp.content}
    Should Be Equal As Strings    ${resp.status_code}    201

Test Get Marker Returns Single Record
    [Tags]  backup  cassandra
    ${resp}=    GET On Session    markersession    /api/v1/data-validation/marker    expected_status=200
    Log    ${resp.content}
    Should Be Equal As Strings    ${resp.status_code}    200
    Should Contain    ${resp.content}    marker
    Should Not Contain    ${resp.content}    Debug:

Test Set Marker Replaces Existing
    [Tags]  backup  cassandra
    ${first_value}=    Set Variable    my-backup/2026-01-01T00:00:00Z
    ${body}=    Set Variable    {"marker": "${first_value}"}
    POST On Session    markersession    /api/v1/data-validation/marker    data=${body}    headers=${headers}    expected_status=201

    ${second_value}=    Set Variable    my-backup/2026-07-07T17:15:00Z
    ${body}=    Set Variable    {"marker": "${second_value}"}
    POST On Session    markersession    /api/v1/data-validation/marker    data=${body}    headers=${headers}    expected_status=201

    ${resp}=    GET On Session    markersession    /api/v1/data-validation/marker   expected_status=200
    Log    ${resp.content}
    Should Contain    ${resp.content}    ${second_value}
    Should Not Contain    ${resp.content}    ${first_value}

Test Get Marker Response Contains Only Marker Value
    [Tags]  backup  cassandra
    ${marker_value}=    Set Variable    my-backup/2026-07-07T17:15:00Z
    ${body}=    Set Variable    {"marker": "${marker_value}"}
    POST On Session    markersession    /api/v1/data-validation/marker    data=${body}    headers=${headers}    expected_status=201

    ${resp}=    GET On Session    markersession    /api/v1/data-validation/marker    expected_status=200
    Log    ${resp.content}
    Should Contain    ${resp.content}    ${marker_value}
    Should Not Contain    ${resp.content}    Debug:
    Should Not Contain    ${resp.content}    host=
