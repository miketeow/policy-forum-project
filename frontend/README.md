## Roadmap
- [ ] Data Lifecycle (Edit / Soft Delete): Right now, if a user makes a typo in a comment or posts something inappropriate, it's there forever. You need basic lifecycle management.
- [ ] Data Bounding (Pagination & Sorting): This is critical. You cannot serve infinite arrays of JSON to your Next.js frontend or your future LLM.
- [ ] Discovery (Search & Filtering): A forum lives and dies by discoverability. Searching via standard SQL ILIKE versus Full-Text Search (FTS).
- [ ] Engagement (Upvotes/Downvotes): This requires a new relational table and introduces interesting concurrency challenges (handling race conditions when two people vote at the same millisecond).
- [ ] LLM Integration: Now, your LLM has bounded data, clean endpoints, and search capabilities to act as tools.
