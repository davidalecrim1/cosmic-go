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

## Chapter 01: Domain Modeling
```
So it is in the mundane world of business. The terminology used by business stakeholders represents a distilled understanding of the domain model, where complex ideas and processes are boiled down to a single word or phrase.
``` 

DDD is a fancy word for requirements engineering in the software world. Undestand and detail the business, then go technical.

```
The name of our unit test describes the behavior that we want to see from the system, and the names of the classes and variables that we use are taken from the business jargon. We could show this code to our nontechnical coworkers, and they would agree that this correctly describes the behavior of the system.
```

I really like this, to sum up with **Clean Code**:
- **Unit Test Name:** Business behaviour
- **Classes:** Business jargon (nouns)
- **Methods:** Business jargon (verbs)

```
This is the part of your code that is closest to the business, the most likely to change, and the place where you deliver the most value to the business. Make it easy to understand and modify.
```

I really can see this in the domain model I've crafted in Go. The book shows how to make it easy to be understood. All domain model should be like this.

## Chapter 02: Repository Pattern

The main value of repository pattern is to implement the dependency inversion principle decoupling our core logic from infrastructure concerns (database).

```
Is this ports and adapters? Or is it hexagonal architecture? Is that the same as onion architecture? What about the clean architecture? What’s a port, and what’s an adapter? Why do you people have so many words for the same thing?

Although some people like to nitpick over the differences, all these are pretty much names for the same thing, and they all boil down to the dependency inversion principle: high-level modules (the domain) should not depend on low-level ones (the infrastructure).
```

This is a great quote.

```
Building fakes for your abstractions is an excellent way to get design feedback: if it’s hard to fake, the abstraction is probably too complicated.
```

Nice quote. I agree with this.

```
If your app is just a simple CRUD (create-read-update-delete) wrapper around a database, then you don’t need a domain model or a repository.
```

DDD is not a silver bullet. Know when to use the ORM in the domain layer and ship your application fast.

## Chapter 03: Interlude on Coupling and Abstractions

```
When we’re unable to change component A for fear of breaking component B, we say that the components have become coupled
```

```
This is the problem with the Ball of Mud pattern: as the application grows, if we’re unable to prevent coupling between elements that have no cohesion, that coupling increases superlinearly until we are no longer able to effectively change our systems.
```

What I really like about this quote is that tests well the developers understand if they are turning the system into a Ball of Mud, because if it is getting harder to test, or worse, not new tests for new features, means the system is going downwards.

```
When we have to tackle a problem from first principles, we usually try to write a simple implementation and then refactor toward better design. We’ll use this approach throughout the book, because it’s how we write code in the real world: start with a solution to the smallest part of the problem, and then iteratively make the solution richer and better designed.
```

Don't be afraid to write bad code, just write it, then stare at it and refactor until you are satisfied.


```
Instead, we like to clearly identify the responsibilities in our codebase, and to separate those responsibilities into small, focused objects that are easy to replace with a test double.
```

Also a lesson, prefer to create your own fakes then mocking without dependency injection.

```
Designing for testability really means designing for extensibility. We trade off a little more complexity for a cleaner design that admits novel use cases.
```

I do believe is a tradeoff worth most of the times.


### Mocks versus Fakes
- **Mocks** are used to verify how something gets used; they have methods like `assert_called_once_with()`
- **Fakes** are working implementations of the thing they’re replacing, but they’re designed for use only in tests. They wouldn’t work "in real life"; our in-memory repository is a good example. 

**Wrap Up**
- Separate the what from the how


### Deep Dive on Mocks, Fakes, Spies and Stubs

#### 1. **Mocks**:
   - **Definition**: Mocks are objects that are used to **verify behavior**. They allow you to check whether certain methods were called with specific arguments during the test. 
   - **Purpose**: Mocks are primarily used for **interaction-based testing**, where the focus is on the interactions between objects rather than the outcome of a function.
   - **How They Work**: Mocks typically record information about how they were called, such as what methods were invoked and with what parameters. After the test, you can assert that the correct methods were called the correct number of times.
   - **Example Use**: When testing whether a service calls an external API with the expected arguments.
   
   ```python
   mock_api = Mock()
   service.do_something()
   mock_api.call.assert_called_once_with(expected_argument)
   ```

#### 2. **Stubs**:
   - **Definition**: Stubs are objects that provide **predefined responses** to calls during a test. Unlike mocks, they do not track interactions; they only supply the necessary return values to keep the test going.
   - **Purpose**: Stubs are used for **state-based testing** where the goal is to test the state of the system after the method has run, without focusing on interactions between objects.
   - **How They Work**: Stubs simply return hardcoded values when their methods are called.
   - **Example Use**: When testing a service that depends on a third-party API, a stub can return the expected API response without making an actual network call.
   
   ```python
   def api_stub():
       return {"data": "expected_response"}
   
   result = service.process(api_stub())
   assert result == expected_outcome
   ```

#### 3. **Fakes**:
   - **Definition**: Fakes are fully functioning implementations, but they are **simpler** or **in-memory versions** of a real system component. They usually have no external dependencies.
   - **Purpose**: Fakes are useful when you need more realistic behavior than what stubs provide but still want to avoid using real dependencies like a database or an external service.
   - **How They Work**: Fakes are often used to simulate real-world services or components, like an in-memory database or a simplified version of an API.
   - **Example Use**: You might use an in-memory database as a fake when testing repository logic, instead of connecting to a real database.
   
   ```python
   class InMemoryRepository:
       def __init__(self):
           self.data = []
       
       def save(self, item):
           self.data.append(item)
   
   fake_repo = InMemoryRepository()
   service.save_item(item, fake_repo)
   assert item in fake_repo.data
   ```

#### 4. **Spies**:
   - **Definition**: Spies are similar to mocks but are more focused on **observing interactions** that happen during the test. They store information about how methods were called, but they are usually more passive, being used for **post-test verification** rather than setting expectations beforehand.
   - **Purpose**: Spies are used to verify interactions and state changes without needing to predefine expectations like mocks do. You can inspect what happened after the test completes.
   - **How They Work**: Like mocks, spies record information about the method calls they receive, but they don’t enforce behavior during the test. You inspect them after the test has run.
   - **Example Use**: You might use a spy to check how many times a method was called or with what parameters, without pre-defining those expectations.
   
   ```python
   class SpyRepository:
       def __init__(self):
           self.saved_items = []
       
       def save(self, item):
           self.saved_items.append(item)
   
   spy_repo = SpyRepository()
   service.save_item(item, spy_repo)
   assert len(spy_repo.saved_items) == 1
   ```

#### Summary of Their Uses:
- **Mocks**: Used to verify **behavior** and interactions (did this method get called with these arguments?).
- **Stubs**: Used to provide **predetermined responses** to method calls (simulate a return value).
- **Fakes**: Provide a **working implementation** that is simpler than the real component (like an in-memory database).
- **Spies**: Capture and record method calls for **post-test inspection** (was this method called, and how?).


## Chapter 04: Service Layer and API

