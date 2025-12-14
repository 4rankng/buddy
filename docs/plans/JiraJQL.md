To retrieve your Jira tickets via the API, you cannot simply use the browser URL you provided. Instead, you need to use the Jira REST API endpoint search.

Here is exactly how to translate your browser view into an API request.

1. The API Endpoint
For Jira Cloud (which gxsbank.atlassian.net is), the standard search endpoint is: GET /rest/api/3/search

2. Your JQL Query
The JQL from your link is:

SQL

project = TSE AND assignee IN (currentUser()) AND status NOT IN (Completed, Closed) ORDER BY created DESC
3. Making the Request (Examples)
You must use Basic Authentication (Email + API Token) for Jira Cloud. You cannot use your login password.

Option A: Using curl (Command Line)
You need to URL-encode the JQL query (spaces become %20, etc.).

Bash

curl --request GET \
  --url 'https://gxsbank.atlassian.net/rest/api/3/search?jql=project%20%3D%20TSE%20AND%20assignee%20IN%20(currentUser())%20AND%20status%20NOT%20IN%20(Completed%2C%20Closed)%20ORDER%20BY%20created%20DESC' \
  --user 'your_email@gxs.com.sg:YOUR_API_TOKEN' \
  --header 'Accept: application/json'
Replace your_email@gxs.com.sg and YOUR_API_TOKEN with your actual details.

Option B: Using Python
This is often easier because it handles the encoding for you.

Python

import requests
from requests.auth import HTTPBasicAuth
import json

url = "https://gxsbank.atlassian.net/rest/api/3/search"

auth = HTTPBasicAuth("your_email@gxs.com.sg", "YOUR_API_TOKEN")

headers = {
  "Accept": "application/json"
}

query = {
  "jql": "project = TSE AND assignee IN (currentUser()) AND status NOT IN (Completed, Closed) ORDER BY created DESC"
}

response = requests.request(
   "GET", 
   url, 
   headers=headers, 
   auth=auth, 
   params=query
)

print(json.dumps(json.loads(response.text), sort_keys=True, indent=4, separators=(",", ": ")))
4. Important Setup Steps
Get an API Token:

Go to https://id.atlassian.com/manage-profile/security/api-tokens.

Click Create API token.

Use this token string as the password in your script/request.

currentUser():

This function relies on the user making the API call. The API token must belong to the user whose "open issues" you want to find. If you want to find tickets for a different user, replace currentUser() with their specific accountId.

Selected Issue:

Your URL ended with &selectedIssue=TSE-815. The search API returns a list of issues. If you only want the details for that one specific ticket, use this endpoint instead:

GET https://gxsbank.atlassian.net/rest/api/3/issue/TSE-815