#!/usr/bin/env python3
"""
Script to fix Swagger documentation metadata after bee generate docs.
This script:
1. Adds proper metadata (title, description, version, schemes)
2. Cleans up tag names from full package paths to simple names
"""

import json
import yaml
import re
import sys

def fix_swagger_metadata(swagger_data):
    """Fix the metadata section of swagger documentation."""
    swagger_data['info'] = {
        "title": "Casdoor RESTful API",
        "description": "Swagger Docs of Casdoor Backend API",
        "version": "1.503.0",
        "contact": {
            "email": "casbin@googlegroups.com"
        }
    }
    
    # Add schemes if not present
    if 'schemes' not in swagger_data or not swagger_data['schemes']:
        swagger_data['schemes'] = ["https", "http"]
    
    return swagger_data

def clean_tag_name(tag):
    """
    Clean up tag names from full package paths to simple readable names.
    Examples:
    - 'github.com/casdoor/casdoor/controllersApiController' -> 'api'
    - 'github.com/casdoor/casdoor/controllersRootController' -> 'root'
    """
    # If it's already a simple tag, keep it
    if '/' not in tag and '.' not in tag:
        return tag
    
    # Extract controller name from package path
    if 'ApiController' in tag:
        return 'api'
    elif 'RootController' in tag:
        return 'root'
    else:
        # For other cases, try to extract a meaningful name
        # Remove package path and Controller suffix
        simple_name = tag.split('/')[-1].replace('Controller', '')
        # Convert from CamelCase to lowercase with spaces
        simple_name = re.sub(r'(?<!^)(?=[A-Z])', ' ', simple_name).lower()
        return simple_name.strip()

def fix_tags_in_paths(swagger_data):
    """Fix tag names in all API paths."""
    if 'paths' not in swagger_data:
        return swagger_data
    
    tag_mapping = {}
    
    for path, methods in swagger_data['paths'].items():
        for method, details in methods.items():
            if 'tags' in details and details['tags']:
                old_tags = details['tags']
                new_tags = []
                for old_tag in old_tags:
                    new_tag = clean_tag_name(old_tag)
                    new_tags.append(new_tag)
                    if old_tag != new_tag:
                        tag_mapping[old_tag] = new_tag
                details['tags'] = new_tags
    
    # Update tags section if it exists
    if 'tags' in swagger_data:
        new_tags = []
        for tag_obj in swagger_data['tags']:
            if 'name' in tag_obj:
                old_name = tag_obj['name']
                tag_obj['name'] = clean_tag_name(old_name)
            new_tags.append(tag_obj)
        swagger_data['tags'] = new_tags
    
    return swagger_data

def remove_html_breaks(swagger_data):
    """Remove <br> tags from descriptions that bee added."""
    if 'paths' not in swagger_data:
        return swagger_data
    
    for path, methods in swagger_data['paths'].items():
        for method, details in methods.items():
            if 'description' in details:
                details['description'] = details['description'].replace('\n<br>', '').replace('<br>', '')
    
    return swagger_data

def fix_operation_ids_and_schemas(swagger_data):
    """
    Fix incorrect operation IDs and response schemas that bee sometimes generates.
    Some payment endpoints are incorrectly labeled as verification endpoints.
    """
    if 'paths' not in swagger_data:
        return swagger_data
    
    # Map of path to correct operation ID and response schema
    fixes = {
        '/api/get-payment': {
            'operationId': 'ApiController.GetPayment',
            'response_schema': '#/definitions/object.Payment'
        },
        '/api/get-payments': {
            'operationId': 'ApiController.GetPayments',
            'response_schema_array': '#/definitions/object.Payment'
        },
        '/api/get-user-payments': {
            'operationId': 'ApiController.GetUserPayments',
            'response_schema_array': '#/definitions/object.Payment'
        },
    }
    
    for path, fix_data in fixes.items():
        if path in swagger_data['paths']:
            for method, details in swagger_data['paths'][path].items():
                # Fix operation ID
                if 'operationId' in details and 'operationId' in fix_data:
                    details['operationId'] = fix_data['operationId']
                
                # Fix response schema
                if 'responses' in details and '200' in details['responses']:
                    response = details['responses']['200']
                    if 'schema' in response:
                        # Fix single object schema
                        if 'response_schema' in fix_data:
                            if '$ref' in response['schema']:
                                response['schema']['$ref'] = fix_data['response_schema']
                        # Fix array schema
                        elif 'response_schema_array' in fix_data:
                            if 'type' in response['schema'] and response['schema']['type'] == 'array':
                                if 'items' in response['schema'] and '$ref' in response['schema']['items']:
                                    response['schema']['items']['$ref'] = fix_data['response_schema_array']
    
    return swagger_data

def main():
    # Load swagger.json
    with open('swagger/swagger.json', 'r', encoding='utf-8') as f:
        swagger_json = json.load(f)
    
    # Fix swagger.json
    swagger_json = fix_swagger_metadata(swagger_json)
    swagger_json = fix_tags_in_paths(swagger_json)
    swagger_json = remove_html_breaks(swagger_json)
    swagger_json = fix_operation_ids_and_schemas(swagger_json)
    
    # Save swagger.json
    with open('swagger/swagger.json', 'w', encoding='utf-8') as f:
        json.dump(swagger_json, f, indent=4, ensure_ascii=False)
    
    print("✓ Fixed swagger/swagger.json")
    
    # Load swagger.yml
    with open('swagger/swagger.yml', 'r', encoding='utf-8') as f:
        swagger_yaml = yaml.safe_load(f)
    
    # Fix swagger.yml
    swagger_yaml = fix_swagger_metadata(swagger_yaml)
    swagger_yaml = fix_tags_in_paths(swagger_yaml)
    swagger_yaml = remove_html_breaks(swagger_yaml)
    swagger_yaml = fix_operation_ids_and_schemas(swagger_yaml)
    
    # Save swagger.yml
    with open('swagger/swagger.yml', 'w', encoding='utf-8') as f:
        yaml.dump(swagger_yaml, f, default_flow_style=False, allow_unicode=True, sort_keys=False)
    
    print("✓ Fixed swagger/swagger.yml")
    print("\nSwagger documentation has been updated successfully!")

if __name__ == '__main__':
    main()
