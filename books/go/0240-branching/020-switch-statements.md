Title: Switch Statements
Id: 4377
Score: 2
Body:
A simple `switch` statement:
```
switch a + b {
case c:
    // do something
case d:
    // do something else
default:
    // do something entirely different
}
```
The above example is equivalent to:
```
if a + b == c {
    // do something
} else if a + b == d {
    // do something else
} else {
    // do something entirely different
}
```


----------


The `default` clause is optional and will be executed if and only if none of the cases compare true, even if it does not appear last, which is acceptable.  The following is semantically the same as the first example:
```
switch a + b {
default:
    // do something entirely different
case c:
    // do something
case d:
    // do something else
}
```
This could be useful if you intend to use the `fallthrough` statement in the `default` clause, which must be the last statement in a case and causes program execution to proceed to the next case:
```
switch a + b {
default:
    // do something entirely different, but then also do something
    fallthrough
case c:
    // do something
case d:
    // do something else
}
```


----------


An empty switch expression is implicitly `true`:
```
switch {
case a + b == c:
    // do something
case a + b == d:
    // do something else
}
```


----------


Switch statements support a simple statement similar to `if` statements:
```
switch n := getNumber(); n {
case 1:
    // do something
case 2:
    // do something else
}
```


----------


Cases can be combined in a comma-separated list if they share the same logic:
```
switch a + b {
case c, d:
    // do something
default:
    // do something entirely different
}
```
|======|