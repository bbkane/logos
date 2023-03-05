# https://github.com/charmbracelet/vhs

Require go
Require bat

Output demo/demo.gif

Set Shell zsh
Set PlaybackSpeed 0.5
Set TypingSpeed 75ms

Set FontSize 22
Set Height 650
Set Width 1050
Set Padding 50

# https://github.com/charmbracelet/vhs/issues/39#issuecomment-1296334921
Hide
Sleep 5s
Show

Type "rm /tmp/testlog.jsonl"
Sleep 200ms
Enter
Sleep 1s

Type "go run ./demo"
Sleep 200ms
Enter
Sleep 1s

Type "bat /tmp/testlog.jsonl"
Sleep 200ms
Enter
Sleep 1s

Sleep 3s