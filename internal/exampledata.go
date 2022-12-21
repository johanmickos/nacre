package nacre

var exampleData [][]byte = [][]byte{
	[]byte("\n"),
	[]byte("\u001b[33;1mhelp) Welcome to nacre's example script!\u001b[0m\n"),
	[]byte("\u001b[33;1mhelp) This script will periodically print out sample data and eventually exit.\u001b[0m\n"),
	[]byte("\n"),
	[]byte("+curl https://example.com/v/users/59f85080-1223-4618-b26c-b5fb6a7f8e2f | jq\n"),
	[]byte("\u001b[1;39m{\n"),
	[]byte("  \u001b[0m\u001b[34;1m\"_id\"\u001b[0m\u001b[1;39m: \u001b[0m\u001b[0;32m\"63a340c802c4f3fe27163893\"\u001b[0m\u001b[1;39m,\n"),
	[]byte("  \u001b[0m\u001b[34;1m\"guid\"\u001b[0m\u001b[1;39m: \u001b[0m\u001b[0;32m\"59f85080-1223-4618-b26c-b5fb6a7f8e2f\"\u001b[0m\u001b[1;39m,\n"),
	[]byte("  \u001b[0m\u001b[34;1m\"isActive\"\u001b[0m\u001b[1;39m: \u001b[0m\u001b[0;39mfalse\u001b[0m\u001b[1;39m,\n"),
	[]byte("  \u001b[0m\u001b[34;1m\"picture\"\u001b[0m\u001b[1;39m: \u001b[0m\u001b[0;32m\"http://placehold.it/32x32\"\u001b[0m\u001b[1;39m,\n"),
	[]byte("  \u001b[0m\u001b[34;1m\"age\"\u001b[0m\u001b[1;39m: \u001b[0m\u001b[0;39m36\u001b[0m\u001b[1;39m,\n"),
	[]byte("  \u001b[0m\u001b[34;1m\"eyeColor\"\u001b[0m\u001b[1;39m: \u001b[0m\u001b[0;32m\"green\"\u001b[0m\u001b[1;39m,\n"),
	[]byte("  \u001b[0m\u001b[34;1m\"name\"\u001b[0m\u001b[1;39m: \u001b[0m\u001b[0;32m\"Maxine Leach\"\u001b[0m\u001b[1;39m,\n"),
	[]byte("  \u001b[0m\u001b[34;1m\"gender\"\u001b[0m\u001b[1;39m: \u001b[0m\u001b[0;32m\"female\"\u001b[0m\u001b[1;39m,\n"),
	[]byte("  \u001b[0m\u001b[34;1m\"company\"\u001b[0m\u001b[1;39m: \u001b[0m\u001b[0;32m\"INVENTURE\"\u001b[0m\u001b[1;39m,\n"),
	[]byte("  \u001b[0m\u001b[34;1m\"email\"\u001b[0m\u001b[1;39m: \u001b[0m\u001b[0;32m\"maxineleach@inventure.com\"\u001b[0m\u001b[1;39m,\n"),
	[]byte("  \u001b[0m\u001b[34;1m\"tags\"\u001b[0m\u001b[1;39m: \u001b[0m\u001b[1;39m[\n"),
	[]byte("    \u001b[0;32m\"magna\"\u001b[0m\u001b[1;39m,\n"),
	[]byte("    \u001b[0;32m\"labore\"\u001b[0m\u001b[1;39m\n"),
	[]byte("  \u001b[1;39m]\u001b[0m\u001b[1;39m,\n"),
	[]byte("  \u001b[0m\u001b[34;1m\"friends\"\u001b[0m\u001b[1;39m: \u001b[0m\u001b[1;39m[\n"),
	[]byte("    \u001b[1;39m{\n"),
	[]byte("      \u001b[0m\u001b[34;1m\"id\"\u001b[0m\u001b[1;39m: \u001b[0m\u001b[0;39m0\u001b[0m\u001b[1;39m,\n"),
	[]byte("      \u001b[0m\u001b[34;1m\"name\"\u001b[0m\u001b[1;39m: \u001b[0m\u001b[0;32m\"Mcconnell Richmond\"\u001b[0m\u001b[1;39m\n"),
	[]byte("    \u001b[1;39m}\u001b[0m\u001b[1;39m\n"),
	[]byte("  \u001b[1;39m]\u001b[0m\u001b[1;39m\n"),
	[]byte("\u001b[1;39m}\u001b[0m\n"),
	[]byte("Bootstrapping example directory structure\n"),
	[]byte("+ mktemp -d /tmp/nacre-XXXX\n"),
	[]byte("+ workdir=/tmp/nacre-myTd\n"),
	[]byte("+ mkdir /tmp/nacre-myTd/bin /tmp/nacre-myTd/examples\n"),
	[]byte("+ touch /tmp/nacre-myTd/README.md /tmp/nacre-myTd/run.sh\n"),
	[]byte("+ ln -s /tmp/nacre-myTd/README.md /tmp/nacre-myTd/symlink-README.md\n"),
	[]byte("+ chmod +x /tmp/nacre-myTd/run.sh\n"),
	[]byte("+ sleep 0.5\n"),
	[]byte("+ ls -lath --color /tmp/nacre-myTd\n"),
	[]byte("total 24K\n"),
	[]byte("drwx------   4 nacre-user nacre-user 4.0K Dec 21 13:22 \u001b[0m\u001b[01;34m.\u001b[0m\n"),
	[]byte("-rw-rw-r--   1 nacre-user nacre-user    0 Dec 21 13:22 README.md\n"),
	[]byte("-rwxrwxr-x   1 nacre-user nacre-user    0 Dec 21 13:22 \u001b[01;32mrun.sh\u001b[0m\n"),
	[]byte("lrwxrwxrwx   1 nacre-user nacre-user   25 Dec 21 13:22 \u001b[01;36msymlink-README.md\u001b[0m -> /tmp/nacre-myTd/README.md\n"),
	[]byte("drwxrwxrwt 165 root       root        12K Dec 21 13:22 \u001b[30;42m..\u001b[0m\n"),
	[]byte("drwxrwxr-x   2 nacre-user nacre-user 4.0K Dec 21 13:22 \u001b[01;34mbin\u001b[0m\n"),
	[]byte("drwxrwxr-x   2 nacre-user nacre-user 4.0K Dec 21 13:22 \u001b[01;34mexamples\u001b[0m\n"),
	[]byte("+ rm -r /tmp/nacre-myTd\n"),
	[]byte("+ set +x\n"),
	[]byte("\n"),
	[]byte("\u001b[33;1mhelp) ✔️  Nacre's example script completed\n\u001b[0m\n"),
}