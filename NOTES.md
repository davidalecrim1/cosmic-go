# Notes from The Book

## Preface
```
In the Python world, we often quote the Zen of Python: "There should be one—​and preferably only one—​obvious way to do it."[1] Unfortunately, as project size grows, the most obvious way of doing things isn’t always the way that helps you manage complexity and evolving requirements.
```

I really like this quote, doesn't matter the language, the right way to do it should be obvious.

## Intro

```
This is so common that software engineers have their own term for chaos: the Big Ball of Mud antipattern (A real-life dependency diagram (source: "Enterprise Dependency: Big Ball of Yarn" by Alex Papadimoulis)).
```

```
The term encapsulation covers two closely related ideas: simplifying behavior and hiding data. In this discussion, we’re using the first sense. We encapsulate behavior by identifying a task that needs to be done in our code and giving that task to a well-defined object or function. We call that object or function an abstraction.
```

```In a traditional OO language like Java or C#, you might use an abstract base class (ABC) or an interface to define an abstraction. In Python you can (and we sometimes do) use ABCs, but you can also happily rely on duck typing.
```

I like this for Python. Relying on duck typing makes so much sense to code our code Pythonic. Now, if one is using Go or another statically typed language, then we can really on inheritance or interfaces (this is better with Go using composition over inheritance).

### The "D" in SOLID
```
High-level modules should not depend on low-level modules. Both should depend on abstractions.
Abstractions should not depend on details. Instead, details should depend on abstractions.

So the first part of the DIP says that our business code (high level modules) shouldn’t depend on technical details(HTTP, OS, TCP, SMTP); instead, both should use abstractions.
```

```
All problems in computer science can be solved by adding another level of indirection.
— David Wheeler
```

This is so true...

```
A Place for All Our Business Logic: The Domain Model
```

This is a nice way of thinking. "Is this a business rule? Do I care about it or the business does? If so, this should be in the core/application/domain of your software. Let's give it a cute nickname: high level modules".

### About Domain Model

```
We’ve found that many developers, when asked to design a new system, will immediately start to build a database schema, with the object model treated as an afterthought. This is where it all starts to go wrong. Instead, behavior should come first and drive our storage requirements. After all, our customers don’t care about the data model. They care about what the system does; otherwise they’d just use a spreadsheet.
```

This is a nice way of thinking. We should see storage from the within (domain model), from the business language, to the low level details.

```
The domain is a fancy way of saying the problem you’re trying to solve.

A model is a map of a process or phenomenon that captures a useful property. 
```