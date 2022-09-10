# retry

```bash
#!/bin/bash

# Retries a command on failure.
# $1 - the max number of attempts
# $2... - the command to run
retry() {
    local -r -i max_attempts="$1"; shift
    local -i attempt_num=1
    until "$@"
    do
        if ((attempt_num==max_attempts))
        then
            echo "\n[RETRY] '$@' failed after $attempt_num attempts\n"
            return 1
        else
            echo "\n[RETRY] '$@' failed attempt $attempt_num, will retry.\n"
            ((attempt_num++))
        fi
    done
}
```

```bash
retry 5 gocov test ./... | gocov report
```
