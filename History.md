# 1.1.3 / 2024-04-28

  * Fix memory leak as of unclosed loggers

# 1.1.2 / 2024-04-20

  * Fix: Cleanup context after checks have run

# 1.1.1 / 2024-04-19

  * Update dependencies

# 1.1.0 / 2024-03-31

  * Modernize code, fix linter errors
  * Replace build system
  * Update dependencies

# 1.0.0 / 2022-05-02

  * Breaking: Remove prometheus support
  * Add go1.18 dependency management
  * Remove old dep management / vendoring

# 0.6.1 / 2018-04-27

  * Update build image
  * Update docs
  * Fix license file

# 0.6.0 / 2018-04-27

  * Update deps, switch to dep for vendoring
  * Deprecate parameter 'warn-only', update yaml unmarshal

# 0.5.2 / 2017-07-23

  * Replace gobuilder as build tool
  * Replace stub with full license text
  * Update README badges

# 0.5.1 / 2016-11-30

  * Fix: Give the process kill a bit more time to succeed

# 0.5.0 / 2016-11-29

  * Add line prefixing to see which check logs lines
  * Log failing checks

# 0.4.1 / 2016-10-11

  * Push builds to GH releases

# 0.4.0 / 2016-08-04

  * `stderr` output will be sent to `stderr` of the program all the time which helps debugging failing commands
  * `stdout` can be attached to `stdout` of the program when `--verbose` flag is passed
  * All commands are now executed using `bash -e -o pipefail -c '<command>'` to make them more reliable and let them fail fast in case of an error

# 0.3.0 / 2016-07-22

  * make check interval and config refresh interval configurable
  * enforce checks to quit before they are started again
  * support URLs as check source

# 0.2.1 / 2016-06-06

  * Vendored dependencies

# 0.2.0 / 2016-06-06

  * Expose metrics about checks for prometheus
  * **Breaking change:** The configuration format changed during this release!

# 0.1.1 / 2016-06-03

  * Bake in the version when using gobuilder

# 0.1.0 / 2016-06-03

  * Added documentation
  * Initital version
