#!/usr/bin/env python3
"""
Autograder Script - Simple Stub Implementation
This script receives a student submission and returns a dummy score if the file exists.

Usage: python3 grade.py <submission_path> <original_filename> <work_dir>
"""

import sys
import os
import json
import time

def main():
    if len(sys.argv) != 4:
        print(json.dumps({
            "error": "Usage: python3 grade.py <submission_path> <original_filename> <work_dir>"
        }))
        sys.exit(1)
    
    submission_path = sys.argv[1]
    original_filename = sys.argv[2] 
    work_dir = sys.argv[3]
    
    # Initialize grading result
    result = {
        "score": 0.0,
        "max_score": 100.0,
        "feedback": "",
        "error": ""
    }
    
    try:
        print(f"🔬 Grading {original_filename}...", file=sys.stderr)
        
        # Check if submission file exists
        if not os.path.exists(submission_path):
            result["error"] = f"Submission file not found: {submission_path}"
        else:
            # Get file size
            file_size = os.path.getsize(submission_path)
            
            # Simulate some grading time
            time.sleep(1)
            
            # Return dummy score and feedback
            result["score"] = 85.0
            result["feedback"] = f"Submission received: {original_filename}\n"
            result["feedback"] += f"File size: {file_size} bytes\n"
            result["feedback"] += "✅ Dummy grading complete - file processed successfully!"
            
            print(f"✅ Grading complete. Score: {result['score']}/{result['max_score']}", file=sys.stderr)
        
    except Exception as e:
        result["error"] = f"Grading failed: {str(e)}"
        print(f"❌ Grading error: {e}", file=sys.stderr)
    
    # Output JSON result to stdout (Go app will parse this)
    print(json.dumps(result, indent=2))

if __name__ == "__main__":
    main()
