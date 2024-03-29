#!/bin/bash
# Simple benchmark which compares speed with other tools

if [ ! -f "./repeater" ]; then
  go build .
fi

max_am=$1

if [ -z "$max_am" ]; then
  max_am=10000
fi

repeater_for_loop() {
  local n=$1
  shift
  local cmd="$@"
  for i in $(seq $n); do
    /bin/bash -c "$cmd" > /dev/null 
  done
}

# Clump up xargs and parallell since they share interface
run_xargs_parallel() {
  xargs_or_parallel=$1
  rep_cmd="$2"
  if command -v $xargs_or_parallel &> /dev/null; then
    printf "gnu $xargs_or_parallel $i processes"
    time printf "$rep_cmd" | $xargs_or_parallel -I {} -P $i /bin/sh -c "{}" > ./p_out_${i}
    # Diff to ensure output is the same = same operations have been done
    diff_check=$(diff ./r_out_${i} ./p_out_${i})
    rm ./p_out_${i}
    if [ -n "$diff_check" ]; then
      printf "found diffs between output, aborting benchmark. Diffs:\n$diff_check"
      rm ./r_out_${i} 
      exit 1
    fi
  fi
}

shift
cmd="$@"

if [ -z "$cmd" ]; then
  cmd="ls"
fi

printf "=== Benchmark start. Am runs: $max_am, Cmd: '$cmd' ==="
echo
for ((i = 10; i <= $max_am; i="${i}0")); do
  echo "--------
== am runs: $i"
  # Uncomment if you want it to be a slow comparison of synchronous loop
  # printf "bash for loop"
  # time repeater_for_loop $i $cmd 
  printf "repeater $i workers"
  time ./repeater -n $i -statistics=false -file="r_out_${i}" -output=FILE -w $i $cmd  > /dev/null
  printf -v rep_cmd '%*s' "$i"
  rep_cmd=$(printf '%s' "${rep_cmd// /${cmd}"\n"}")

  run_xargs_parallel "parallel" "$rep_cmd" 
  run_xargs_parallel "xargs" "$rep_cmd"
  
  # Gargs (https://github.com/brentp/gargs)
  if command -v gargs &> /dev/null; then
    printf "gargs $i processes"
    time printf "$rep_cmd" | gargs --procs $i "{}" > ./s_out_${i}
    diff_check=$(diff ./r_out_${i} ./s_out_${i})
    rm ./s_out_${i}
    if [ -n "$diff_check" ]; then
      printf "found diffs between output, aborting benchmark. Diffs:\n$diff_check"
      rm ./r_out_${i} 
      exit 1
    fi
  fi

  rm ./r_out_${i} 
done


# Clean up the binary
rm repeater
