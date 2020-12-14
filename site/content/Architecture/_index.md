---
title: "Architecture"
linkTitle: "Architecture"
weight: 4
description: >
    How is VHS implemented internally?
---

## Introduction
`vhs` is designed for flexibility and operates on the concept of a data flow that originates with a [source](#sources)
and terminates with one or more [sinks](#sinks). Sources may capture network data, read data from files, etc. Sinks may
write data to cloud or local storage, standard output, or send data to other destinations. Along the way, data may pass
through a series of input [modifiers](#input-modifiers) and [formats](#input-formats) and output
[modifiers](#output-modifiers) and [formats](#output-formats) that transform the data.

More architectural details coming soon.