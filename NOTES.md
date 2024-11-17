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
```
The first is an **application service (our service layer)**. Its job is to handle requests from the outside world and to orchestrate an operation. The second type of service is a domain service. This is the name for a piece of logic that belongs in the domain model but doesn’t sit naturally inside a stateful entity or value object.
```

We want to have fast unit tests and keep the integration and e2e as minimal as possible.

**Domain Service** = business logic that doesn't fit the entity or value object. e.g. calculating a tax
**Application Service** = orchestration and probably persistence in the database. e.g. creating a user

The **application services** are like a API with our use cases to our domain, therefore we could refactor our domain as we see fit without breaking the API to the external world (or REST API in the handler or presentation layer).


## Chapter 05: TDD in High Gear and Low Gear

```
Let’s see what happens if we take this a step further. Since we can test our software against the service layer, we don’t really need tests for the domain model anymore. Instead, we could rewrite all of the domain-level tests from [chapter_01_domain_model] in terms of the service layer

Why would we want to do that?

Tests are supposed to help us change our system fearlessly, but often we see teams writing too many tests against their domain model. This causes problems when they come to change their codebase and find that they need to update tens or even hundreds of unit tests.
```

This is interesting.

```
Every line of code that we put in a test is like a blob of glue, holding the system in a particular shape. The more low-level tests we have, the harder it will be to change things.
```

This is important. When we need to change our systems, we also need to change our tests.

```
Extreme programming (XP) exhorts us to "listen to the code." When we’re writing tests, we might find that the code is hard to use or notice a code smell. This is a trigger for us to refactor, and to reconsider our design.
```

### High and Low Gear
 
```
Most of the time, when we are adding a new feature or fixing a bug, we don’t need to make extensive changes to the domain model. In these cases, we prefer to write tests against services because of the lower coupling and higher coverage.
```

I've decided not to swap the domain structs for primitives given I rather use structs. Even if the domain is coupled with the outside, I believe is a worph trade off. Reviewing what Uncle Bob says about this with ChatGPT:
- Passing primitive types (like strings, integers, etc.) to achieve decoupling can reduce direct dependencies between layers, but it also runs the risk of losing some of the clarity and expressiveness that comes from using structured domain objects. Domain objects typically encapsulate related data and business logic, providing better readability and preventing errors related to passing incorrect or misaligned primitives.
- Passing many arguments is discouraged, whether they are primitives or structured objects. Instead, Uncle Bob would suggest bundling related data into cohesive objects, making function calls cleaner.
- Decoupling is encouraged, but not at the cost of clarity or maintainability. The service layer should ideally depend on abstractions (e.g., interfaces or **DTOs**), which can allow decoupling without relying too heavily on primitives.

I could use DTOs, but I do not see why in small software and a language like Go that focus on simplicity.


```
Extreme programming (XP) exhorts us to "listen to the code." When we’re writing tests, we might find that the code is hard to use or notice a code smell. This is a trigger for us to refactor, and to reconsider our design.
```

I do enjoy this. Remember why to TDD. You are a professional (quoting Uncle Bob from Clean Coder). It's your responsability to ship code to production as you were doing the most important thing in the world.

The E2E tests are great because you have high coverage and make sure large scale changes don't break your software.

```
Because the tests are written in the domain language, they act as living documentation for our model. A new team member can read these tests to quickly understand how the system works and how the core concepts interrelate.

We often "sketch" new behaviors by writing tests at this level to see how the code might look. When we want to improve the design of the code, though, we will need to replace or delete these tests, because they are tightly coupled to a particular implementation.

Most of the time, when we are adding a new feature or fixing a bug, we don’t need to make extensive changes to the domain model. In these cases, we prefer to write tests against services because of the lower coupling and higher coverage.
```

This is great for unit tests on the domain.

```
We still have direct dependencies on the domain in our service-layer tests, because we use domain objects to set up our test data and to invoke our service-layer functions.

To have a service layer that’s fully decoupled from the domain, we need to rewrite its API to work in terms of primitives.
```

I disagree with this and don't see it making sense in Go. Also, thinking about Clean Code, using the domain objects make the code more readable and expressive. I do like that and believe it's a worth trade off.


```
In general, if you find yourself needing to do domain-layer stuff directly in your service-layer tests, it may be an indication that your service layer is incomplete.
```

This does seem like a nice tip.

```
Write the bulk of your tests against the service layer
These edge-to-edge tests offer a good trade-off between coverage, runtime, and efficiency. Each test tends to cover one code path of a feature and use fakes for I/O. This is the place to exhaustively cover all the edge cases and the ins and outs of your business logic.[1]

Maintain a small core of tests written against your domain model
These tests have highly focused coverage and are more brittle, but they have the highest feedback. Don’t be afraid to delete these tests if the functionality is later covered by tests at the service layer.
```

The service layer should be the core of unit testing. Think about when you need to do "low gear" and create some tests over the domain, and when you need to use "high gear" and focus on service layer testing.


## Chapter 06: Unit of Work Pattern
```
If the Repository pattern is our abstraction over the idea of persistent storage, the Unit of Work (UoW) pattern is our abstraction over the idea of atomic operations. It will allow us to finally and fully decouple our service layer from the data layer.
```

In Go the Unit of Work doesn't seem to make sense it is in the book. With `pgxpool` we can reuse the connections on the database, and the hard thing is orchestrate transactions. I will create the [UoW based on this article](https://threedots.tech/post/database-transactions-in-go/).


## Chapter 07: Aggregates and Consistency Boundaries
```
Adding the Product aggregate shows a preview of where we’re headed: we’ll introduce a new model object called Product to wrap multiple batches, and we’ll make the old allocate() domain service available as a method on Product instead.
```

### Invariants, Concurrency, and Locks
```
We usually solve this problem by applying locks to our database tables. This prevents two operations from happening simultaneously on the same row or same table.

As we start to think about scaling up our app, we realize that our model of allocating lines against all available batches may not scale. If we process tens of thousands of orders per hour, and hundreds of thousands of order lines, we can’t hold a lock over the whole batches table for every single one—​we’ll get deadlocks or performance problems at the very least.
```

This is the classic optimistic locking versus pessimist locking.

```
The Aggregate pattern is a design pattern from the DDD community that helps us to resolve this tension. An aggregate is just a domain object that contains other domain objects and lets us treat the whole collection as a single unit.

The only way to modify the objects inside the aggregate is to load the whole thing, and to call methods on the aggregate itself.

An AGGREGATE is a cluster of associated objects that we treat as a unit for the purpose of data changes. - Eric Evans on DDD.

Per Evans, our aggregate has a root entity (the Cart) that encapsulates access to items. Each item has its own identity, but other parts of the system will always refer to the Cart only as an indivisible whole.
```

### Choosing an Aggregate
```
What aggregate should we use for our system? The choice is somewhat arbitrary, but it’s important. The aggregate will be the boundary where we make sure every operation ends in a consistent state. This helps us to reason about our software and prevent weird race issues. We want to draw a boundary around a small number of objects—the smaller, the better, for performance—that have to be consistent with one another, and we need to give this boundary a good name.
```

This DDD thinking in this chapter is great. It's worth a review and practice.

```
This Product might not look like what you’d expect a Product model to look like. No price, no description, no dimensions. Our allocation service doesn’t care about any of those things. This is the power of bounded contexts; the concept of a product in one app can be very different from another. See the following sidebar for more discussion.

In our example, the allocation service has Product(sku, batches), whereas the ecommerce will have Product(sku, description, price, image_url, dimensions, etc…​). As a rule of thumb, your domain models should include only the data that they need for performing calculations.
```

### One Aggregate = One Repository
```
The rule that repositories should only return aggregates is the main place where we enforce the convention that aggregates are the only way into our domain model. Be wary of breaking it!
```

I see in this chapter that there is not "right" way to model a domain. You can have two developer create distinct domains and if both are performant and scalable, that's fine. That also means there will be domain models poorly designed. Maybe because the absense of knowledge by the developer or thinking it throw with TDD.

```
Version numbers are just one way to implement optimistic locking. You could achieve the same thing by setting the Postgres transaction isolation level to SERIALIZABLE, but that often comes at a severe performance cost. Version numbers also make implicit concepts explicit.
```

```
Pessimistic concurrency control works under the assumption that two users are going to cause conflicts, and we want to prevent conflicts in all cases, so we lock everything just to be safe. In our example, that would mean locking the whole batches table, or using SELECT FOR UPDATE—we’re pretending that we’ve ruled those out for performance reasons, but in real life you’d want to do some evaluations and measurements of your own.

With pessimistic locking, you don’t need to think about handling failures because the database will prevent them for you (although you do need to think about deadlocks). With optimistic locking, you need to explicitly handle the possibility of failures in the (hopefully unlikely) case of a clash.
```

I myself believe that SELECT FOR UPDATE works fine in most cases. When I was implementing this, the version_id seems faulty and can be messed up be the developer.

### Some Thoughts on Using ORM

The same pain I had for writing SQL statements I had on defining and getting the ORM to work. I myself like the idea of using raw SQL, but I do believe the ORM is worth my attetion and some more tests in the future. I've needed to add a BeforeSave hook on the ORM to make sure the deallocation works. I'm not familiar if this if an issue only with GORM, or most ORMs.

Later I found out that I can preload the data, and use the feature `FullSaveAssociations` to do that automatically.

GORM has a way to handle deletions automatically when updating associated slices. This feature leverages the Association mode, where GORM manages the changes between the current and new state of the associations, including deletions.

To delete items automatically from an associated slice that have been removed, you need to use the Select or Omit methods when saving your model. By default, GORM does not automatically delete removed associations to prevent accidental data loss. Instead, you explicitly need to tell GORM to update the association.

### Wrap Up
```
Choosing the right aggregate is key, and it’s a decision you may revisit over time. You can read more about it in multiple DDD books. We also recommend these three online papers on effective aggregate design by Vaughn Vernon (the "red book" author).
```

```
At the risk of laboring the point—​we’ve been at pains to point out that each pattern comes at a cost. Each layer of indirection has a price in terms of complexity and duplication in our code and will be confusing to programmers who’ve never seen these patterns before. If your app is essentially a simple CRUD wrapper around a database and isn’t likely to be anything more than that in the foreseeable future, you don’t need these patterns. Go ahead and use Django, and save yourself a lot of bother.
```

I agree with this. Simple CRUD should be simple, the repository and UoW patterns with this TDD comes with costs. If we respect the Dependency Inversion Principle and our code is easy to test, then it's fine.


## Chapter 08: Events and the Message Bus

```
Faced with this requirement, many teams reach for microservices integrated via HTTP APIs. But if they’re not careful, they’ll end up producing the most chaotic mess of all: the distributed big ball of mud.
```

```
The requirement "Try to allocate some stock, and send an email if it fails" is an example of workflow orchestration: it’s a set of steps that the system has to follow to achieve a goal.
```

It's a nice thing to think. Orchestration layer (applicarion or service layer, what ever you call) it's the place for things like that.

```
Rule of thumb: if you can’t describe what your function does without using words like "then" or "and," you might be violating the SRP.
```

```
One formulation of the SRP is that each class should have only a single reason to change. When we switch from email to SMS, we shouldn’t have to update our allocate() function, because that’s clearly a separate responsibility.
```

```
We’re actually addressing a code smell we had until now, which is that we were using exceptions for control flow. In general, if you’re implementing domain events, don’t raise exceptions to describe the same domain concept. As you’ll see later when we handle events in the Unit of Work pattern, it’s confusing to have to reason about events and exceptions together.
```

```
A message bus basically says, "When I see this event, I should invoke the following handler function." In other words, it’s a simple publish-subscribe system. Handlers are subscribed to receive events, which we publish to the bus. 
```

```
Domain events give us a way to handle workflows in our system. We often find, listening to our domain experts, that they express requirements in a causal or temporal way—for example, "When we try to allocate stock but there’s none available, then we should send an email to the buying team.

The magic words "When X, then Y" often tell us about an event that we can make concrete in our system. Treating events as first-class things in our model helps us make our code more testable and observable, and it helps isolate concerns.
```

For the Domain Events in this code, I didn't liked Option 3 using Go, it would overcomplicate with the seen method in the UoW. Therefore I decided to use Option 2 and make service layer save the events as orchestraton layer.

## Chapter 09: Going to Town on the Message Bus

```
In this chapter, we’ll start to make events more fundamental to the internal structure of our application. We’ll move from the current state in Before: the message bus is an optional add-on, where events are an optional side effect to the situation in The message bus is now the main entrypoint to the service layer, where everything goes via the message bus, and our app has been transformed fundamentally into a message processor.
```

```
What Have We Achieved?
Events are simple dataclasses that define the data structures for inputs and internal messages within our system. This is quite powerful from a DDD standpoint, since events often translate really well into business language (look up event storming if you haven’t already).

Handlers are the way we react to events. They can call down to our model or call out to external services. We can define multiple handlers for a single event if we want to. Handlers can also raise other events. This allows us to be very granular about what a handler does and really stick to the SRP.
```

This might be a good architecture depending on the business goal with the software.


```
Our ongoing objective with these architectural patterns is to try to have the complexity of our application grow more slowly than its size. When we go all in on the message bus, as always we pay a price in terms of architectural complexity (see Whole app is a message bus: the trade-offs), but we buy ourselves a pattern that can handle almost arbitrarily complex requirements without needing any further conceptual or architectural change to the way we do things.

Our ongoing objective with these architectural patterns is to try to have the complexity of our application grow more slowly than its size. When we go all in on the message bus, as always we pay a price in terms of architectural complexity (see Whole app is a message bus: the trade-offs), but we buy ourselves a pattern that can handle almost arbitrarily complex requirements without needing any further conceptual or architectural change to the way we do things.
```

This offers some reflection. In order to avoid complexity, we may create more complexity that we might not know if it is needed. Think about this when choosing an architecture. The goal is always to focus on simplicity, and when the requirements (i.e. the business goal) are complex, we design a architecture that can hold that complexity without technical debt.

## Chapter 10: Commands and Command Handler

```
Like events, commands are a type of message—​instructions sent by one part of a system to another. We usually represent commands with dumb data structures and can handle them in much the same way as events.

Commands are sent by one actor to another specific actor with the expectation that a particular thing will happen as a result. When we post a form to an API handler, we are sending a command. We name commands with imperative mood verb phrases like "allocate stock" or "delay shipment."

Commands capture intent. They express our wish for the system to do something. As a result, when they fail, the sender needs to receive error information.

Events are broadcast by an actor to all interested listeners. When we publish BatchQuantityChanged, we don’t know who’s going to pick it up. We name events with past-tense verb phrases like "order allocated to stock" or "shipment delayed."

We often use events to spread the knowledge about successful commands.

Events capture facts about things that happened in the past. Since we don’t know who’s handling an event, senders should not care whether the receivers succeeded or failed. Events versus commands recaps the differences.
```

Events = Past
Command = Imperative, Do This Now!

```
Events go to a dispatcher that can delegate to multiple handlers per event.
It catches and logs errors but doesn’t let them interrupt message processing.
```

```
The command dispatcher expects just one handler per command.
If any errors are raised, they fail fast and will bubble up.
```

```
Retrying operations that might fail is probably the single best way to improve the resilience of our software. Again, the Unit of Work and Command Handler patterns mean that each attempt starts from a consistent state and won’t leave things half-finished.
```

## Chapter 11: Event-Driven Architecture: Using Events to Integrate Microservices

```
Before we get into that, let’s talk about the alternatives. We regularly talk to engineers who are trying to build out a microservices architecture. Often they are migrating from an existing application, and their first instinct is to split their system into nouns.

What nouns have we introduced so far in our system? Well, we have batches of stock, orders, products, and customers. So a naive attempt at breaking up the system might have looked like Context diagram with noun-based services (notice that we’ve named our system after a noun, Batches, instead of Allocation).
```

```
This style of architecture, where we create a microservice per database table and treat our HTTP APIs as CRUD interfaces to anemic models, is the most common initial way for people to approach service-oriented design.

This works fine for systems that are very simple, but it can quickly degrade into a distributed ball of mud.
```

```
When two things have to be changed together, we say that they are coupled. We can think of this failure cascade as a kind of temporal coupling: every part of the system has to work at the same time for any part of it to work. As the system gets bigger, there is an exponentially increasing probability that some part is degraded.

We can never completely avoid coupling, except by having our software not talk to any other software. What we want is to avoid inappropriate coupling. Connascence provides a mental model for understanding the strength and type of coupling inherent in different architectural styles. Read all about it at connascence.io.
```

```
How do we get appropriate coupling? We’ve already seen part of the answer, which is that we should think in terms of verbs, not nouns. Our domain model is about modeling a business process. It’s not a static data model about a thing; it’s a model of a verb.

So instead of thinking about a system for orders and a system for batches, we think about a system for ordering and a system for allocating, and so on.

When we separate things this way, it’s a little easier to see which system should be responsible for what. When thinking about ordering, really we want to make sure that when we place an order, the order is placed. Everything else can happen later, so long as it happens.

Like aggregates, microservices should be consistency boundaries. Between two services, we can accept eventual consistency, and that means we don’t need to rely on synchronous calls. Each service accepts commands from the outside world and raises events to record the result. Other services can listen to those events to trigger the next steps in the workflow.
```

This is nice.

```
Why is this better? First, because things can fail independently, it’s easier to handle degraded behavior: we can still take orders if the allocation system is having a bad day.
```

Also in deployment and the responsible squad for each microservice.

### Internal Versus External Events
```
It’s a good idea to keep the distinction between internal and external events clear. Some events may come from the outside, and some events may get upgraded and published externally, but not all of them will. This is particularly important if you get into event sourcing (very much a topic for another book, though).
```

## Chapter 12: Command-Query Responsibility Segregation (CQRS)
```
In this chapter, we’re going to start with a fairly uncontroversial insight: reads (queries) and writes (commands) are different, so they should be treated differently (or have their responsibilities segregated, if you will). Then we’re going to push that insight as far as we can.
```

This is an interesting perspective.

```
As soon as we render the product page, the data is already stale. This insight is key to understanding why reads can be safely inconsistent: we’ll always need to check the current state of our system when we come to allocate, because all distributed systems are inconsistent. As soon as you have a web server and two customers, you have the potential for stale data.

No matter what we do, we’re always going to find that our software systems are inconsistent with reality, and so we’ll always need business processes to cope with these edge cases. It’s OK to trade performance for consistency on the read side, because stale data is essentially unavoidable.
```

Nice perspective on consistency on distribuited systems.

### Post/Redirect/Get and CQS
```
This technique is a simple example of command-query separation (CQS).[1] We follow one simple rule: functions should either modify state or answer questions, but never both. This makes software easier to reason about: we should always be able to ask, "Are the lights on?" without flicking the light switch.

When building APIs, we can apply the same design technique by returning a 201 Created, or a 202 Accepted, with a Location header containing the URI of our new resources. What’s important here isn’t the status code we use but the logical separation of work into a write phase and a query phase.
```

Based on this, I have created a new layer on the software for queries (i.e. reads), this based on CQRS should have it's own domain logic and operation handling. Because here it's simple, we are just using a thin http layer with the database instead of the UoW, Message Bus and Service Layer.

### Your Domain Model Is Not Optimized for Read Operations
```
What we’re seeing here are the effects of having a domain model that is designed primarily for write operations, while our requirements for reads are often conceptually quite different.

This is the chin-stroking-architect’s justification for CQRS. As we’ve said before, a domain model is not a data model—​we’re trying to capture the way the business works: workflow, rules around state changes, messages exchanged; concerns about how the system reacts to external events and user input. Most of this stuff is totally irrelevant for read-only operations.

This justification for CQRS is related to the justification for the Domain Model pattern. If you’re building a simple CRUD app, reads and writes are going to be closely related, so you don’t need a domain model or CQRS. But the more complex your domain, the more likely you are to need both.
```

Remember that CQRS is for complex domains.

```
But is that actually any easier to write or understand than the raw SQL version from the code example in Hold On to Your Lunch, Folks? It may not look too bad up there, but we can tell you it took several attempts, and plenty of digging through the SQLAlchemy docs. SQL is just SQL.
```

I do like the idea of not relying too much on the ORM.

### SELECT N+1 and Other Performance Considerations
```
The so-called SELECT N+1 problem is a common performance problem with ORMs: when retrieving a list of objects, your ORM will often perform an initial query to, say, get all the IDs of the objects it needs, and then issue individual queries for each object to retrieve their attributes. This is especially likely if there are any foreign-key relationships on your objects.
```

If your "product" table has relationships, the ORM will get all the objects, then get each single object again with the details of the ORM. Some ORM already have features to help this.

```
Beyond SELECT N+1, you may have other reasons for wanting to decouple the way you persist state changes from the way that you retrieve current state. A set of fully normalized relational tables is a good way to make sure that write operations never cause data corruption. But retrieving data using lots of joins can be slow. It’s common in such cases to add some denormalized views, build read replicas, or even add caching layers.
```

I do agree with this.

```
Because read replicas can be inconsistent, there’s no limit to how many we can have. If you’re struggling to scale a system with a complex data store, ask whether you could build a simpler read model.

Keeping the read model up to date is the challenge! Database views (materialized or otherwise) and triggers are a common solution, but that limits you to your database. We’d like to show you how to reuse our event-driven architecture instead.
```

I've decided not do to the Denormalized Table as in the book, because with Go I don't see the UoW fitting with direct interaction with the database. Based on what I've built, I would need a Read Repository, but it seems quite simplier to just dependency inject the database on the HTTP layer for read operations. I don't see this mix of UoW (given it is for write operations) and Repository. It would only make sense to create a UoW and Repository, but given Go is statically typed, that means creating also a new domain for reads. This is too much for such a simple read operation.

Based on all the options of the chapter, the best one I see with Go is using a Denormalized View.

```
Often, your read operations will be acting on the same conceptual objects as your write model, so using the ORM, adding some read methods to your repositories, and using domain model classes for your read operations is just fine.
```

Remember CQRS is for **COMPLEX** domain. It is almost as breaking the application into READ and WRITE, with each using it's own database. I would say to everyone, **KISS** (keep it simple stupid!). Don't overengineer something because it looks cool. Know if you really need it. Study about it. Make an MVP with the new approach.

```
In our book example, the read operations act on quite different conceptual entities to our domain model. The allocation service thinks in terms of Batches for a single SKU, but users care about allocations for a whole order, with multiple SKUs, so using the ORM ends up being a little awkward. We’d be quite tempted to go with the raw-SQL view we showed right at the beginning of the chapter.
```

Maybe this is also a domain issue with the book. The domain could be improved to allow both Reads and Writes instead of working nice with writes and HELL with reads.

## Epilogue

```
Aggregates are a consistency boundary. In general, each use case should update a single aggregate at a time. One handler fetches one aggregate from a repository, modifies its state, and raises any events that happen as a result. If you need data from another part of the system, it’s totally fine to use a read model, but avoid updating multiple aggregates in a single transaction. When we choose to separate code into different aggregates, we’re explicitly choosing to make them eventually consistent with one another.

Bidirectional links are often a sign that your aggregates aren’t right. In our original code, a Document knew about its containing Folder, and the Folder had a collection of Documents. This makes it easy to traverse the object graph but stops us from thinking properly about the consistency boundaries we need. We break apart aggregates by using references instead. In the new model, a Document had reference to its parent_folder but had no way to directly access the Folder.
```

```
If we needed to read data, we avoided writing complex loops and transforms and tried to replace them with straight SQL. For example, one of our screens was a tree view of folders and documents.
```

Remember how to use ORM and SQL. It's not one or the other.

```
We recommend domain modeling as a first step. In many overgrown systems, the engineers, product owners, and customers no longer speak the same language. Business stakeholders speak about the system in abstract, process-focused terms, while developers are forced to speak about the system as it physically exists in its wild and chaotic state.
```

