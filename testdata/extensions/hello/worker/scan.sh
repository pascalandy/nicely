# scan.sh <file> reads a file line by line and prints its progress on
# stdout, as an encoder such as HandBrakeCLI does. Its exit code is its only
# answer, so hello scan reads none of its stdout.
lines=0
while IFS= read -r line || [ -n "$line" ]; do
	lines=$((lines + 1))
	printf 'Scanning: line %d\n' "$lines"
done <"$1" || exit 1
