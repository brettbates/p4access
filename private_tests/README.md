# Private Tests
To help find edge-cases from real data whilst keeping this project open source, this test harness can be used to run 

> p4 -u a.user access read|write path

against a local broker pointing at your Perforce server. I recommend a local server so it doesn't affect the real broker.

By default the runner looks for a p4broker called 'localhost:1998', you can set this with P4PORT variable at the top of ./run_tests.py

For the broker config, make sure it points to your test p4 server and add the stanza:

```
command: access
{
    action=filter;
    execute="/path/to/p4access/p4access";
}
```

Then set the standard [env variables](../README.md) and run the broker.

# Tests
In this folder, create a file called 'private.csv'. 

> This is ignored in the .gitignore and it should stay that way, don't submit any private data back to this repo.

The fields are:

|  User.name  | Requested Access |           Path            |          Expected           |
| :---------: | :--------------: | :-----------------------: | :-------------------------: |
| brett.bates |       read       | //a/path/to/somewhere/... | A_group_read&&A_group_read2 |
| brett.bates |      write       | //a/path/to/somewhere/... |            NONE             |

which is like so in csv:
```csv
brett.bates,read,//a/path/to/somwhere/...,A_group_read1&&A_group_read2
```

Options for the expected column:

| Expected | Comment                                  |
| :------- | :--------------------------------------- |
| X        | Expect there to be group x only returned |
| X&&Y     | Expect groups X and Y in that order      |
| NONE     | Expect no groups to be returned          |
| ERROR    | Expect an error to be thrown             |
