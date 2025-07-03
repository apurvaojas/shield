# Sample OPA Policy for Demo Application
package demo.authz

import future.keywords.if
import future.keywords.in

# Default deny
default allow := false

# Allow if user has the required role for the resource
allow if {
    # Get user roles for this application
    user_roles := input.user.roles
    
    # Check if user has required role
    required_role := resource_role_mapping[input.resource]
    required_role in user_roles
    
    # Additional checks
    valid_action
    within_business_hours
}

# Define resource to role mapping
resource_role_mapping := {
    "/api/users": "admin",
    "/api/reports": "viewer",
    "/api/settings": "admin",
    "/api/data": "editor"
}

# Check if action is valid for the resource
valid_action if {
    input.action == "read"
}

valid_action if {
    input.action == "write"
    input.user.roles[_] in ["admin", "editor"]
}

valid_action if {
    input.action == "delete"
    input.user.roles[_] == "admin"
}

# Business hours check (optional)
within_business_hours if {
    # For demo purposes, always allow
    true
}

within_business_hours if {
    # Real implementation would check time
    hour := time.clock([time.now_ns()])[0]
    hour >= 9
    hour <= 17
}
