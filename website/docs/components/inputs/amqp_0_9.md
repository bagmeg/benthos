---
title: amqp_0_9
type: input
status: stable
categories: ["Services"]
---

<!--
     THIS FILE IS AUTOGENERATED!

     To make changes please edit the contents of:
     lib/input/amqp_0_9.go
-->

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';


Connects to an AMQP (0.91) queue. AMQP is a messaging protocol used by various
message brokers, including RabbitMQ.


<Tabs defaultValue="common" values={[
  { label: 'Common', value: 'common', },
  { label: 'Advanced', value: 'advanced', },
]}>

<TabItem value="common">

```yml
# Common config fields, showing default values
input:
  label: ""
  amqp_0_9:
    urls: []
    queue: ""
    consumer_tag: ""
    prefetch_count: 10
```

</TabItem>
<TabItem value="advanced">

```yml
# All config fields, showing default values
input:
  label: ""
  amqp_0_9:
    urls: []
    queue: ""
    queue_declare:
      enabled: false
      durable: true
    bindings_declare: []
    consumer_tag: ""
    auto_ack: false
    nack_reject_patterns: []
    prefetch_count: 10
    prefetch_size: 0
    tls:
      enabled: false
      skip_cert_verify: false
      enable_renegotiation: false
      root_cas: ""
      root_cas_file: ""
      client_certs: []
```

</TabItem>
</Tabs>

TLS is automatic when connecting to an `amqps` URL, but custom
settings can be enabled in the `tls` section.

### Metadata

This input adds the following metadata fields to each message:

``` text
- amqp_content_type
- amqp_content_encoding
- amqp_delivery_mode
- amqp_priority
- amqp_correlation_id
- amqp_reply_to
- amqp_expiration
- amqp_message_id
- amqp_timestamp
- amqp_type
- amqp_user_id
- amqp_app_id
- amqp_consumer_tag
- amqp_delivery_tag
- amqp_redelivered
- amqp_exchange
- amqp_routing_key
- All existing message headers, including nested headers prefixed with the key of their respective parent.
```

You can access these metadata fields using
[function interpolation](/docs/configuration/interpolation#metadata).

## Fields

### `urls`

A list of URLs to connect to. The first URL to successfully establish a connection will be used until the connection is closed. If an item of the list contains commas it will be expanded into multiple URLs.


Type: `array`  
Default: `[]`  
Requires version 3.58.0 or newer  

```yml
# Examples

urls:
  - amqp://guest:guest@127.0.0.1:5672/

urls:
  - amqp://127.0.0.1:5672/,amqp://127.0.0.2:5672/

urls:
  - amqp://127.0.0.1:5672/
  - amqp://127.0.0.2:5672/
```

### `queue`

An AMQP queue to consume from.


Type: `string`  
Default: `""`  

### `queue_declare`

Allows you to passively declare the target queue. If the queue already exists
then the declaration passively verifies that they match the target fields.


Type: `object`  

### `queue_declare.enabled`

Whether to enable queue declaration.


Type: `bool`  
Default: `false`  

### `queue_declare.durable`

Whether the declared queue is durable.


Type: `bool`  
Default: `true`  

### `bindings_declare`

Allows you to passively declare bindings for the target queue.


Type: `array`  
Default: `[]`  

```yml
# Examples

bindings_declare:
  - exchange: foo
    key: bar
```

### `bindings_declare[].exchange`

The exchange of the declared binding.


Type: `string`  
Default: `""`  

### `bindings_declare[].key`

The key of the declared binding.


Type: `string`  
Default: `""`  

### `consumer_tag`

A consumer tag.


Type: `string`  
Default: `""`  

### `auto_ack`

Acknowledge messages automatically as they are consumed rather than waiting for acknowledgments from downstream. This can improve throughput and prevent the pipeline from blocking but at the cost of eliminating delivery guarantees.


Type: `bool`  
Default: `false`  

### `nack_reject_patterns`

A list of regular expression patterns whereby if a message that has failed to be delivered by Benthos has an error that matches it will be dropped (or delivered to a dead-letter queue if one exists). By default failed messages are nacked with requeue enabled.


Type: `array`  
Default: `[]`  
Requires version 3.64.0 or newer  

```yml
# Examples

nack_reject_patterns:
  - ^reject me please:.+$
```

### `prefetch_count`

The maximum number of pending messages to have consumed at a time.


Type: `int`  
Default: `10`  

### `prefetch_size`

The maximum amount of pending messages measured in bytes to have consumed at a time.


Type: `int`  
Default: `0`  

### `tls`

Custom TLS settings can be used to override system defaults.


Type: `object`  

### `tls.enabled`

Whether custom TLS settings are enabled.


Type: `bool`  
Default: `false`  

### `tls.skip_cert_verify`

Whether to skip server side certificate verification.


Type: `bool`  
Default: `false`  

### `tls.enable_renegotiation`

Whether to allow the remote server to repeatedly request renegotiation. Enable this option if you're seeing the error message `local error: tls: no renegotiation`.


Type: `bool`  
Default: `false`  
Requires version 3.45.0 or newer  

### `tls.root_cas`

An optional root certificate authority to use. This is a string, representing a certificate chain from the parent trusted root certificate, to possible intermediate signing certificates, to the host certificate.


Type: `string`  
Default: `""`  

```yml
# Examples

root_cas: |-
  -----BEGIN CERTIFICATE-----
  ...
  -----END CERTIFICATE-----
```

### `tls.root_cas_file`

An optional path of a root certificate authority file to use. This is a file, often with a .pem extension, containing a certificate chain from the parent trusted root certificate, to possible intermediate signing certificates, to the host certificate.


Type: `string`  
Default: `""`  

```yml
# Examples

root_cas_file: ./root_cas.pem
```

### `tls.client_certs`

A list of client certificates to use. For each certificate either the fields `cert` and `key`, or `cert_file` and `key_file` should be specified, but not both.


Type: `array`  

```yml
# Examples

client_certs:
  - cert: foo
    key: bar

client_certs:
  - cert_file: ./example.pem
    key_file: ./example.key
```

### `tls.client_certs[].cert`

A plain text certificate to use.


Type: `string`  
Default: `""`  

### `tls.client_certs[].key`

A plain text certificate key to use.


Type: `string`  
Default: `""`  

### `tls.client_certs[].cert_file`

The path of a certificate to use.


Type: `string`  
Default: `""`  

### `tls.client_certs[].key_file`

The path of a certificate key to use.


Type: `string`  
Default: `""`  

### `tls.client_certs[].password`

A plain text password for when the private key is a password encrypted PEM block according to RFC 1423. Warning: Since it does not authenticate the ciphertext, it is vulnerable to padding oracle attacks that can let an attacker recover the plaintext.


Type: `string`  
Default: `""`  

```yml
# Examples

password: foo

password: ${KEY_PASSWORD}
```


