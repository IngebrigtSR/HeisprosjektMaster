

What is concurrency? What is parallelism? What's the difference?

concurrency is when different tasks run in overlapping time.
paralellism is when tasks are executed in the excact same time (needs multiple cores).

Difference is paralellism is about doing several things at once, 
and concurrency helps deal with several things at once, but not executing things in parallell.


Why have machines become increasingly multicore in the past decade?

To improve performnce without having to increase the clock frequency.


What kinds of problems motivates the need for concurrent execution?
(Or phrased differently: What problems do concurrency help in solving?)

Concurrency is useful in timing problems that frequently occur in real time systems.


Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
(Come back to this after you have worked on part 4 of this exercise)

It makes it possible to do several tasks at once, but it is harder to make the code robust. 
So it opens up more possibilities allowing for more flexible software, but can be more dfficult to program.



What are the differences between processes, threads, green threads, and coroutines?
Your answer here

A process contains all resource needed to run a program, a thread is a part of a process
that can be scheduled for excecution and uses all resources provided by the process.

Green threads are threads that run in a simulated enviroment, allowing threading os hardware that originally does not support multithreading

Coroutines are subroutines that can be suspended and resumed, important for concurrency.


Which one of these do pthread_create() (C/POSIX), threading.Thread() (Python), go (Go) create?

pthread_create() creates a  thread
threading.thread() creates a thread.
go creates a green thread.


How does pythons Global Interpreter Lock (GIL) influence the way a python Thread behaves?

GIL forces concurrency, so threads are not parallell.


With this in mind: What is the workaround for the GIL (Hint: it's another module)?

Use the multiprocessing module to take advantage of multiple processor cores.


What does func GOMAXPROCS(n int) int change?

Limits the number of threads that can can run simultaniously, essencially capping the amount of prossecor cores that can be used in parallell.
