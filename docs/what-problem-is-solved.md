Joplin is an open source note management application that I actively use. Some time ago, it was necessary to edit notes automatically. The task seemed simple, but we managed to collect enough rakes along the way. While I was figuring it out, I wrote a micro client on go. At the same time, I raised the experimental API of the console client in docker and integrated it into the application. I share the information I found and the code.

## What task is being solved

Joplin applications are used on different platforms and support several data synchronization methods. In my case, they are synchronized via S3.

There are several notes with todo lists (for example: work, household chores, shopping lists, hobbies). Each list has its own note. It is necessary to select, sort and display relevant tasks in a special note on a certain basis. That is, some auto-generated todo sheet "for the near future".

The task in general is to periodically receive note data, process and update the content of the selected note.

## Finding a solution

The obvious idea is to use the API. The application documentation and Google pointed to a certain web clipper API. There is not enough information on it. In the CLI application, the server mode is marked experimental feature. The first attempts were unsuccessful, temporarily postponed. OK, what else is there?

There is a Joplin Server: implementation of the synchronization server from the authors of the application. Its API is not public, the server was not intended as an API for third-party clients. Technically, it is possible to apply it. But I don't want to use a non-public API. Maybe there are more options?

Joplin provides the opportunity to create your own clients. First, let's try to create our own client implementation with synchronization in S3.

## Solution 1: go directly to S3

To write a client, you will have to figure out the file formats

### Note file

The structure of the note file is simple: this is a text file.md. Contains the header, body, and metadata. The separator is an empty string:

```
Headline

Content

id: f23f0132f2124bdea9db91c747853434
parent_id: 55f71434e23b412cb87d3b1cbcff823b
<...>
```

When editing, it is important not to forget to update the value of the updated_time, user_updated_time parameters. This way the other clients will pick up the file changes during synchronization.

### Lock file

In addition to editing the note, it is important not to forget that there are other clients performing synchronization at the same time. The lock file is responsible for the consistency of the data. It is a file with the name of the format  
`locks/<lock_type>_<app_type>_<app_id>.json`  
The contents of the file, judging by the documentation, do not matter: the body contains the same data as in the name. The file body is probably legacy. Just in case, we will write the same json object that is written to the lock files in other applications:

```
{"type":2,"clientType":1,"clientId":"f5fe97768f3f4188b062a55619b8753e"}
```

We are interested in write lock, exclusive lock in Joplin terms.

#### Mechanism for taking the lock from the documentation

- Checking for blockages
    
- If there are no locks, create a lock file and go back to checking
    
- If there is a lock, we check whether we own it. If we do not own the lock, we return to the verification
    
- If there is a lock and we own it, we make the change
    
- Removing the lock
    

These two types of files are enough to solve the current task. The client will perform the crown task of updating the note

#### The task of updating the note

- we get a lock for recording
    
- process:  
    
    - we are reading .md files
        
    - filtering by parent_id
        
    - processing the data
        
    - collecting the content of the updated note
        
    - if the data for the update has not changed, exit
        
    - updating the note
        
- removing the lock
    

The result can be found in the repository

## Solution 2: web Clipper API

And if you don't reinvent the ~~bicycle~~ client, and don't touch the files directly? The application has an API, it remains to learn how to cook it. And what exactly is this web clipper API?

The Web Clipper API is an HTTP interface for browser extensions. In the settings of the desktop application, you can launch a web server for integration with the browser. And besides the GUI, Joplin has a CLI version. And in the CLI application, you can also start the server mode. Total: launch the cli, set up synchronization, enable the API. Editing notes via the API.

We are trying to wrap it in a docker. We run the configuration in json format and start in server mode: `joplin server start'. The application is listening to `127.0.0.1:41184'. At startup, you will additionally need socat to proxy to `0.0.0.0'. The sync.interval parameter does not start synchronization for some reason. Ok, just periodically run the cli command `joplin sync'. The API is running and available.

Now it remains to add the http client and the implementation of the joplin provider for the http client to the already created application.

An example implementation and a docker-wrapped API in the repository.  
I have implemented the same interface as for s3. This is not very optimal, but it will do as proof of concept.

## Pros and cons of solutions

**Advantages of working with S3**

- high speed of change delivery: app -> S3 -> target_client
    
- the expansion possibilities are not limited
    
- the synchronization format is quite simple
    

**Disadvantages of working with S3**

- it is difficult to expand the functionality, you will have to deal with logic and formats
    
- one specific protocol is used
    

**Advantages of working with the API**

- no dependence on the note format
    
- there is no dependence on the synchronization protocol
    
- basic functionality has been implemented
    

**Disadvantages of working with the API**

- the speed of change delivery is lower: scheme app -> CLI -> S3 -> target_client
    
- the infrastructure is more complicated: you need to raise an additional service with the API
    
- 2 times more disk space (S3+CLI client)
    
- CLI server experimental option
    
- the possibilities are limited by API methods
    

## Conclusion

As often happens, choosing a solution is a compromise. I chose to work with S3 directly. But it's also good to be able to quickly implement the necessary functionality using the API. I did not check the work of the joplin server API, but I think this integration will also work.

Thanks for your attention!

## Links

[https://joplinapp.org](https://joplinapp.org/)  
https://github.com/laurent22/joplin  
[How lock works in joplin](https://joplinapp.org/help/dev/spec/sync_lock#acquiring-an-exclusive-lock )  
[Joplin web clipper](https://joplinapp.org/help/apps/clipper)
[Original of this article in Russian](http://web.archive.org/web/20240201064639/https://habr.com/ru/articles/785072/)
