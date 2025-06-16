for i in {1..12}; do
  curl -H "X-API-Key: test-key-123" \
       -F "file=@submission.zip" \
       http://localhost:8080/submit
  echo "Request $i - $(date)"
done