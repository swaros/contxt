# anko
<!-- TOC -->- [anko](#anko)
- [anko](#anko)
  - [what is anko](#what-is-anko)
  - [runtime specific](#runtime-specific)
  - [anko language usage](#anko-language-usage)

<!-- /TOC -->

## what is anko
checkout [Anko](https://github.com/mattn/anko) to get a understanding of the language.
it is basically a go script that is executed in a go runtime. in context it is extended with some functions to interact with the context.

by using anko  commands you can write way more specific scripts than you could do with bash. and they are independent from the host system.

## runtime specific

anko commands will be executed before the script. so you can use them to prepare variables for the later script.

```yaml
task:
  - id: example
    cmd:
      - setVar("myvar", "hello world")
    script:
      - echo $myvar
```

even this example makes no sense, it shows how you can use anko to prepare the script.


## anko language usage

