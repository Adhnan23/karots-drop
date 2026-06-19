package main

import (
	"fmt"
	"os"
)

func cmdCompletion(args []string) {
	if len(args) > 0 && args[0] == "zsh" {
		// zsh not implemented yet, fall through to bash
	}
	fmt.Print(bashCompletion)
	if !isCompletionSourced() {
		fmt.Fprintln(os.Stderr, "\n# Run: source <(karots-drop completion)")
	}
}

func isCompletionSourced() bool {
	return false
}

const bashCompletion = `_karots_drop_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    if [[ ${COMP_CWORD} -eq 1 ]]; then
        opts="serve send get clip version health completion"
        COMPREPLY=($(compgen -W "${opts}" -- ${cur}))
        return 0
    fi

    case "${prev}" in
        serve)
            opts="--addr --token --rate-limit --max-items --delete-on-retrieve --ttl --bind"
            ;;
        send)
            opts="--file --server --encrypt --qr --json --ttl --compact"
            ;;
        get)
            opts="--server --json --key"
            ;;
        clip)
            opts="--server --encrypt --qr --json --watch"
            ;;
        *)
            return 0
            ;;
    esac
    COMPREPLY=($(compgen -W "${opts}" -- ${cur}))
}
complete -F _karots_drop_completion karots-drop
`
