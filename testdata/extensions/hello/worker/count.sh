# count.sh <file> prints the number of lines of a file as JSON, such as
# {"lines":2}, with shell builtins alone. A last line without a newline
# counts too.
lines=0
while IFS= read -r line || [ -n "$line" ]; do
	lines=$((lines + 1))
done <"$1" || exit 1
printf '{"lines":%d}\n' "$lines"
