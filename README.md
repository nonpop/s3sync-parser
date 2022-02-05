# s3sync-parser

NOTE: This tool is made for my own use and I do not plan to put *any* extra
effort in it. I probably won't even accept pull requests but you are free to
fork it and modify as needed.

## Description

`aws s3 sync <localPath> s3://<bucket> --delete --dryrun` produces output from
which it is difficult to see which files are being *moved* because a move is
done in two steps (delete old + upload new), and they are not even necessarily
close to each other in the output.

This tool finds moves from the output and shows them as such. It's not 100%
accurate: if you delete a file and upload another file with the same name
(modulo path), it will be considered a move. It also shows Dropbox's cache
separately so it's easier to ignore.

## Usage

Build & install the binary somewhere. Then pipe the result of the aws command to
the tool.
