# Project Specification

In this project, you will implement a key-value store backed by a **copy-on-write trie**.  
Tries are efficient ordered-tree data structures for retrieving a value for a given key.  

To simplify the explanation, we will assume that the **keys are variable-length strings**, but in practice they can be any arbitrary type.

---

## Trie Structure
- Each node in a trie can have multiple child nodes representing different possible next characters.
- The key-value store can store string keys mapped to values of any type.
- The value of a key is stored in the node representing the **last character of that key** (also called the **terminal node**).

---

## Example
Consider inserting the key-value pairs:

- `(1980,"harvard")`
- `(1982, "stanford")`

The trie would look like this:

<img width="747" height="709" alt="image" src="https://github.com/user-attachments/assets/99bf7daf-3449-4e59-915b-94c4b79bd9f3" />

The two keys share the same parent node.  
- The value `harvard` corresponding to key `"8"` is stored in the **left child**.  
- The value `"stanford"` corresponding to key `"2"` is stored in the **right child**.

# Task #1 - Copy-On-Write Trie

In this task, you will need to implement a **copy-on-write trie**.

---

## Copy-On-Write Concept
- Operations do **not directly modify** the nodes of the original trie.  
- Instead:
  1. New nodes are created for the modified data.  
  2. A **new root** is returned for the newly modified trie.  
- Copy-on-write allows the trie to be accessed **after each operation at any time** with **minimal overhead**.

---

## Example
Consider inserting `(1983, "CMU")` into the trie from the previous example:

- We create a **new root node** .  
- We **reuse** the existing child nodes from the original trie.  
- We create a **new value node** for `3`.  

This way:
- The old trie remains intact (immutable).  
- The new trie includes the inserted key-value pair without affecting prior versions.  

<img width="986" height="709" alt="image" src="https://github.com/user-attachments/assets/4d9be986-86f8-48a8-a9e1-b3c01fd9c955" />

## Additional Example

If we then:

Remove `(1980, "harvard")`  

We obtain the following updated trie:

<img width="1191" height="674" alt="image" src="https://github.com/user-attachments/assets/36d624ba-1d6d-4da8-a837-8c27bcbf08d9" />

# Task #2 — Concurrent Key-Value Store (Specification & Design)

Below is a clear, self-contained design and checklist for implementing a **concurrent key-value store** backed by your **copy-on-write trie**. This focuses on the concurrency guarantees you specified: **multiple concurrent readers** and **writers that don't block readers** (readers see consistent snapshots). It also covers lifetime safety for pointers (the `ValueGuard` concept), writer correctness, and practical notes (reclamation, testing, pitfalls).

---

## Goal (restatement)
Implement a thread-safe key-value store with these operations:

- `Get -> ValueGuard`  
- `Insert ` (no return)  
- `Delete` (no return)

**Concurrency properties:**
- Multiple readers run concurrently without blocking each other.
- Readers never block writers; writers never block readers.
- Writes produce new versions (copy-on-write). Readers that hold a `ValueGuard` must remain safe even if the key is deleted or overwritten later.
- The store must not lose updates (concurrent writers handled safely).

---

## Core ideas (summary)
1. **Atomic current root pointer**  
   Maintain a single atomically-swappable pointer to the current trie root (e.g., `atomic.Value` or `atomic.Pointer[trie]`). Readers load it atomically; writers compute a new root then attempt to install it.

2. **Copy-on-write semantics for writers**  
   Writers clone required nodes along the path and build a new root representing the updated trie. They **install** the new root atomically.

3. **ValueGuard (per-Get handle)**  
   `Get` returns a small object (the `ValueGuard`) that contains:
   - a pointer to the value stored inside the trie node (not a stack/local copy),
   - a reference to the root object that was active when the `Get` ran (to pin the snapshot and prevent reclamation).
   Holding `ValueGuard` ensures the value pointer is valid.

4. **Writers use CAS (compare-and-swap)**  
   To avoid lost updates when there are concurrent writers, use CAS semantics: read `oldRoot`, build `newRoot`, `CompareAndSwap(oldRoot, newRoot)`. If CAS fails, retry from the new `oldRoot`.

5. **Rely on Go GC for reclamation**  
   In Go, the `ValueGuard`’s root reference keeps that root reachable; when all guards referencing a root are dropped, the root becomes eligible for GC. No manual refcounting required.



