implement this 

Request URL
https://doorman.infra.prd.g-bank.app/api/rds/dml/create_ticket
Request Method
POST
Status Code
200 OK
Remote Address
10.148.163.34:443

request payload
{
    "accountID": "559634300081",
    "clusterName": "prd-payments-rpp-adapter-rds-mysql",
    "database": "",
    "schema": "rpp_adapter",
    "originalQuery": "-- rpp_cashin_stuck_100_0, update timestamp to resolve optimistic lock\nUPDATE workflow_execution\nSET state = 100,\n    attempt = 1,\n    updated_at = NOW(),\n    data = JSON_SET(data, '$.State', 100)\nWHERE run_id IN ('5c34e6ab0fea334f88b9b4cdb781902f')\nAND workflow_id = 'wf_ct_cashin'\nAND state = 100;\n\n-- rpp_cashin_validation_failed_122_0, retry validation\nUPDATE workflow_execution\nSET state = 100,\n\t  attempt = 1,\n\t  data = JSON_SET(data, '$.State', 100)\nWHERE run_id IN ('240776d1927d3e8ea1871de454d6f8e0')\nAND workflow_id = 'wf_ct_cashin'\nAND state = 122;\n\n",
    "rollbackQuery": "-- rpp_cashin_stuck_100_0_rollback, reset attempt back to 0\nUPDATE workflow_execution\nSET attempt = 0,\n    data = JSON_SET(data, '$.State', 100)\nWHERE run_id IN ('5c34e6ab0fea334f88b9b4cdb781902f')\nAND workflow_id = 'wf_ct_cashin'\nAND state = 100;\n\n-- rpp_cashin_validation_failed_122_0_rollback\nUPDATE workflow_execution\nSET state = 122,\n\t  attempt = 0,\n\t  data = JSON_SET(data, '$.State', 122)\nWHERE run_id IN ('240776d1927d3e8ea1871de454d6f8e0')\nAND workflow_id = 'wf_ct_cashin';\n\n",
    "toolLabel": "direct",
    "skipWhereClause": false,
    "skipRollbackQuery": false,
    "skipRollbackQueryReason": null,
    "note": "TS-4560"
}


response
{
    "code": 200,
    "errors": null,
    "message": null,
    "result": [
        {
            "id": 43198,
            "submitter": "ext.vietdung.nguyen",
            "status": "wait for approval",
            "owners": [
                "azuan.zairein",
                "ext.huang.kun",
                "ext.sathish.kumar",
                "ext.vietdung.nguyen",
                "khorjeng.yong"
            ],
            "oncallUsers": [
                "dummy"
            ],
            "env": "Production",
            "accountID": "559634300081",
            "accountName": "prd-payments",
            "dbsManaged": false,
            "clusterName": "prd-payments-rpp-adapter-rds-mysql",
            "clusterType": "MySQL",
            "clusterID": 112446,
            "instanceName": "prd-payments-rpp-adapter-rds-mysql",
            "instanceID": 7879,
            "oncallGroup": "oncall-dummy",
            "techFamily": "dbmy",
            "pagePath": "https://doorman.myteksi.net/rds/dml",
            "note": "TS-4560",
            "batch": 50000,
            "schema": "rpp_adapter",
            "evaluateRows": 2,
            "affectRows": 2,
            "percentage": 0,
            "originalQuery": "-- rpp_cashin_stuck_100_0, update timestamp to resolve optimistic lock\nUPDATE workflow_execution\nSET state = 100,\n    attempt = 1,\n    updated_at = NOW(),\n    data = JSON_SET(data, '$.State', 100)\nWHERE run_id IN ('5c34e6ab0fea334f88b9b4cdb781902f')\nAND workflow_id = 'wf_ct_cashin'\nAND state = 100;\n\n-- rpp_cashin_validation_failed_122_0, retry validation\nUPDATE workflow_execution\nSET state = 100,\n\t  attempt = 1,\n\t  data = JSON_SET(data, '$.State', 100)\nWHERE run_id IN ('240776d1927d3e8ea1871de454d6f8e0')\nAND workflow_id = 'wf_ct_cashin'\nAND state = 122;\n\n",
            "rollbackQuery": "-- rpp_cashin_stuck_100_0_rollback, reset attempt back to 0\nUPDATE workflow_execution\nSET attempt = 0,\n    data = JSON_SET(data, '$.State', 100)\nWHERE run_id IN ('5c34e6ab0fea334f88b9b4cdb781902f')\nAND workflow_id = 'wf_ct_cashin'\nAND state = 100;\n\n-- rpp_cashin_validation_failed_122_0_rollback\nUPDATE workflow_execution\nSET state = 122,\n\t  attempt = 0,\n\t  data = JSON_SET(data, '$.State', 122)\nWHERE run_id IN ('240776d1927d3e8ea1871de454d6f8e0')\nAND workflow_id = 'wf_ct_cashin';\n\n",
            "subQuery": "",
            "subMinID": 0,
            "subMaxID": 0,
            "encrypted": false,
            "pattern": "UPDATE `workflow_execution` SET `state`=?, `attempt`=?, `updated_at`=NOW(), `data`=JSON_SET(`data`, ?, ?) WHERE `run_id` IN (?) AND `workflow_id` = ? AND `state` = ?\nUPDATE `workflow_execution` SET `state`=?, `attempt`=?, `data`=JSON_SET(`data`, ?, ?) WHERE `run_id` IN (?) AND `workflow_id` = ? AND `state` = ?",
            "eoApprover": "",
            "dbaApprover": "auto_approve",
            "toolLabel": "direct",
            "database": "",
            "fileDir": "prd-payments-rpp-adapter-rds-mysql_",
            "fileType": "",
            "fileSize": 0,
            "pauseLabel": 1,
            "warningMsg": "",
            "remark": "",
            "peakTime": null,
            "archived": false,
            "createdAt": "2026-01-06T08:50:18Z",
            "skipWhereClause": false,
            "skipRollbackQuery": false,
            "skipRollbackQueryReason": ""
        }
    ],
    "requestID": "dc33d197-35fa-491c-8ebb-aae80b23cdb1"
}