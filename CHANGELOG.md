## 1.2
* Upgraded worker pattern to guarantee amount of runs
* Introduced flag retryOnFail
    - If set, the task will be retried if it fails
    - If set to false, the task will simply be started -n amount of times
    - Previous behavoiur were to simply repeat a command -n amonut of times, now the default is to successfully run the task -n amount of times (felt more useful)
* Upgraded progress print to include amount of failures

## 1.1
* Changed default behaviour of result from STDOUT to HIDDEN
* Changed name of REPORT_FILE -> FILE
* Added new flag -result
    - Set this to some filename and it will print a json containing info of each repeated task. This is basically 'progress' file, but once repeat is done
