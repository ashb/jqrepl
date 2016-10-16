# jq-repl

## What is it?

A REPL to make exploring data with JQ easier.

I'm a huge fan of [JQ][jq] and use it in a lot of small utilities or to
explore JSON APIs from the command line, and I often found myself doing things
like this:

```bash
aws ec2 describe-images | jq '.Images'

# Hmm, that's still large
aws ec2 describe-images | jq '.Images | keys'

aws ec2 describe-images | jq '.Images | .Tags'
```

i.e. I was using `jq` as a tool to explore a complex JSON data structure --
but each invokation of `aws ec2 describe-images` took 5 to 15 seconds which
which made the process of building a jq filter quite jaring.

Now, I could have just piped the result of the `aws` command to a file and then
invoked JQ on that many times, and to start with that's what I did. But it
turned out that each of the `Images` above has differente keys, so finding the
error with jq alone was painful, so in another terminal I fired up ipython,
loaded that JSON file into a python dictionary and started exploring the data
that way. Somehow it got suggested that a REPL for JQ would be the right tool
for this job - and thus the seed for this tool was planted. (P.S. Samir and
James: this is all your fault for egging me on)

## Does it work?

**Sort of**. I'm working on it slowly.

I am using this project as excuse and reason to learn Go so it will take me a
while to get it functional and bug free. And then even longer so that other
pepole can use it.


## What's the current state?

```
$ aws ec2 describe-images --owner self | jq-repl
  0 » type
$out[0] = "object"

  1 » keys
$out[1] = [
    "Images"
]

  2 » .Images[0]
$out[2] = {
  "VirtualizationType": "hvm",
  "Name": "leader 2015-11-05T16-50-35Z",
  "Tags": [
    {
      "Value": "2015-11-05T16:50:35Z",
      "Key": "build_date"
    }
  ],
  "Hypervisor": "xen",
  "SriovNetSupport": "simple",
  "ImageId": "ami-abc01234",
  "State": "available",
  "BlockDeviceMappings": [
    {
      "DeviceName": "/dev/sda1",
      "Ebs": {
        "DeleteOnTermination": true,
        "SnapshotId": "snap-01234fed",
        "VolumeSize": 16,
        "VolumeType": "gp2",
        "Encrypted": false
      }
    },
  ],
  "Architecture": "x86_64",
  "RootDeviceType": "ebs",
  "RootDeviceName": "/dev/sda1",
  "CreationDate": "2015-11-05T16:55:15.000Z",
  "Public": false,
  "ImageType": "machine",
  "Description": "My AMI"
}
```

So far all fairly mundane. This is where I think things start to get
interesting - you will be able to refer back to previous results.

```
  3 » $_ | .Name
$out[3] = "leader 2015-11-05T16-50-35Z"
```

### Special variables

We define some special variables.


- `$_`

  The previous result. Equivalent to `$out[-1]`

- `$__`

  The result before last. Equivalent to `$out[-2]`

- `$out`

  An array of all previous results. You can use negative indcies to count
  backwards from the end.

## Current status and future plans


 - [x] Accept JSON input on stdin
 - [ ] Accept JSON input from a file
 - [ ] Sepcify a command to run to get input
 - [ ] Be able to re-execute command to refresh input
 - [ ] Autocompletion
 - [x] Refer to previous results
 - [ ] Build up complex filters by refering to previous results.
 - [ ] Automated tests for the JqRepl methods

# Building it


## Prerequisites

* [JQ souce code][JQ src] and anything it needs to compile

[JQ]: https://stedolan.github.io/jq/
[JQ src]: https://stedolan.github.io/jq/download/

## Build it

It doesn't do much of anything yet. But to build it you will need to do
something like this:

```bash
curl -fL https://github.com/stedolan/jq/releases/download/jq-1.5/jq-1.5.tar.gz | tar -zx
cd jq-1.5
./configure --disable-maintainer-mode --prefix=$PWD/BUILD
# We could run `make install` but we only actually need these components.
make install-libLTLIBRARIES install-includeHEADERS
go test ./...
```

I have no idea if this will work on platforms other than OSX right now. I will
work on that later once I have some basic functionality
