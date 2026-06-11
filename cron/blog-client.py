#!/usr/bin/env python3
import sys
import os
import json
import urllib.request
import urllib.error

# Path configurations
ENV_PATH = "/home/latif/raevtar/.env.production"
API_BASE = "http://localhost:8080/api/v1"

def get_auth_token():
    if not os.path.exists(ENV_PATH):
        print(f"Error: env file not found at {ENV_PATH}", file=sys.stderr)
        sys.exit(1)
    
    with open(ENV_PATH, "r") as f:
        for line in f:
            line = line.strip()
            if line.startswith("RAEVTAR_ADMIN_KEY="):
                return line.split("=", 1)[1].strip('"').strip("'")
                
    print("Error: RAEVTAR_ADMIN_KEY not found in env file", file=sys.stderr)
    sys.exit(1)

def make_request(url, method="GET", data=None, token=None):
    headers = {
        "Content-Type": "application/json"
    }
    if token:
        headers["Authorization"] = f"Bearer {token}"
        
    req_body = None
    if data is not None:
        req_body = json.dumps(data).encode("utf-8")
        
    req = urllib.request.Request(url, method=method, data=req_body, headers=headers)
    
    try:
        with urllib.request.urlopen(req) as response:
            return response.status, json.loads(response.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        err_body = e.read().decode("utf-8")
        try:
            err_json = json.loads(err_body)
            err_msg = err_json.get("error", err_body)
        except json.JSONDecodeError:
            err_msg = err_body
        print(f"HTTP Error {e.code}: {err_msg}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Network error: {str(e)}", file=sys.stderr)
        sys.exit(1)

def get_inbox(token):
    url = f"{API_BASE}/editorial-inbox?ready=true"
    status, res = make_request(url, "GET", token=token)
    print(json.dumps(res, indent=2))

def get_posts():
    url = f"{API_BASE}/posts"
    status, res = make_request(url, "GET")
    print(json.dumps(res, indent=2))

def publish(payload_str, token):
    try:
        data = json.loads(payload_str)
    except json.JSONDecodeError as e:
        print(f"Error: Invalid payload JSON: {str(e)}", file=sys.stderr)
        sys.exit(1)
        
    url = f"{API_BASE}/posts"
    status, res = make_request(url, "POST", data=data, token=token)
    print(json.dumps(res, indent=2))

def done_inbox(item_id, post_id, token):
    # 1. Fetch current item details
    get_url = f"{API_BASE}/editorial-inbox/{item_id}"
    _, item = make_request(get_url, "GET", token=token)
    
    # 2. Prepare update payload conforming to model.EditorialInboxUpdate
    update_payload = {
        "source_type": item["source_type"],
        "source_value": item["source_value"],
        "category_hint": item.get("category_hint", ""),
        "priority": item.get("priority", 0),
        "not_before": item["not_before"],
        "deadline": item.get("deadline"),
        "note": item.get("note", ""),
        "mode": item["mode"],
        "status": "done",
        "published_post_id": int(post_id),
        "failure_note": "",
        "failure_meta": ""
    }
    
    # 3. POST the update back
    post_url = f"{API_BASE}/editorial-inbox/{item_id}"
    status, res = make_request(post_url, "POST", data=update_payload, token=token)
    print(json.dumps(res, indent=2))

def main():
    if len(sys.argv) < 2:
        print("Usage: blog-client.py [get-inbox|get-posts|publish|done-inbox] [args]", file=sys.stderr)
        sys.exit(1)
        
    cmd = sys.argv[1]
    
    if cmd == "get-posts":
        get_posts()
        return
        
    token = get_auth_token()
    
    if cmd == "get-inbox":
        get_inbox(token)
    elif cmd == "publish":
        if len(sys.argv) < 3:
            print("Usage: blog-client.py publish '<json_payload>'", file=sys.stderr)
            sys.exit(1)
        publish(sys.argv[2], token)
    elif cmd == "done-inbox":
        if len(sys.argv) < 4:
            print("Usage: blog-client.py done-inbox <item_id> <post_id>", file=sys.stderr)
            sys.exit(1)
        done_inbox(sys.argv[2], sys.argv[3], token)
    else:
        print(f"Unknown command: {cmd}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
