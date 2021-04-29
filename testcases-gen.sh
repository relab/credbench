#!/bin/sh

# | Courses | Exams per Course | Students |
# |---------|------------------|----------|
# | 10      | 2, 5             | 200      |
# | 100     | 2, 5             | 200      |
# | 1000    | 2                | 100      |

bin=./dist/ctbench
default_flags="--config dev-config.json"

mkdir ./testcases
for c in 10 100 1000; do
  for x in 2 5; do
    for s in 100 200; do
      echo "Generating config for test case: $c-$x-$s"
      ${bin} ${default_flags} test generate case "testcases/test_${c}_${x}_${s}.json" -t 1000 -c $c -x $x -s $s
    done
  done
done