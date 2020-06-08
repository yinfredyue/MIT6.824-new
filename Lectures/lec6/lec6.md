# Lec6

In the distributed system we have seen, there's some point where the system relies on a single entity/authority to make a critical decision. MapReduce has a single master, GFS has a single master, VM-FT has a single test-and-set server. Having a single master makes the management easier and avoids the split brain problem.