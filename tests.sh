


cp config.json config.json.backup

# multi line echo
cat <<EOF > config.json
{
  "port": 8889,

  "graph": {
    "name": "test_graph",
    "path": "./"
  }
}
EOF

# setup notes

mkdir test_graph

cat <<EOF > test_graph/metadata.json
{ "id": "9ec6f641-a2e4-49a8-abac-a9cd34ecfd92" }
EOF

cat <<EOF > test_graph/content.md
# content
EOF




go build

./rat &

sleep 5

failed=0
hurl tests.hurl --color
if [ $? != 0 ]; then
    failed=1
fi

kill -9 $(pgrep ^rat$)

cat logs.log

rm rat

rm config.json

mv config.json.backup config.json

rm -rf test_graph


if [ $failed -eq 1 ]; then
    exit 1
fi
