# here are all the commands paper CLI supports.

```rs
paper help // prints the help text.

paper new . // creates a new paper server in the current location , make sure to include the dot
paper new PATH // creates a new paper server at the given location

paper run // runs paper in the current directory
paper run PATH // runs paper in the selected path

paper delete // deletes everything BUT the paper.jar in the current path , will ask for confirmation by default.
- paper delete --confirm // delete without asking for confirmation

paper delete PATH // deletes everything BUT the paper.jar in the selected path , will ask for confirmation by default.
- paper delete PATH --confirm // delete without asking for confirmation
```