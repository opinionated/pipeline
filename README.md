# pipeline

In order to make our analyzer design as extensible as possible we adopted a modular pipeline approach. Each module scores a set of articles in relation to a single article based on an attribute. These attributes (see Table P.1) then are used to create modules (see Figure P.1). With the data that the articles contain, and are associated with, we are able to effectively score them based on their similarity to other articles, as well as determine affinity for certain specific attributes, all updated in the graph database.
The pipeline is written using basic Golang channel structures to best utilize the language’s built-in concurrency and low memory footprint. The modules themselves are easily extensible, as they are each an interface and can be publicly accessed.


Each module changes the score of an analyzable component. They all work independantly and can connect easily. 

There may be some broken components here.

The module functionality works well, but I haven't written an easy way to test modules yet. 

We should think about reogranizing this a little bit, it could easily get really messy. 

