# Distributed Information Reference (DIR)

A Distributed Information Reference (DIR) is a reference in the
Distributed Information System (DIS). DIS is organized as a tree
of nodes.

```text
           ___
          |___| <------------------- Root Node
      ___/ _|_ \___
     |___||___||___| <-------------- Nodes
           _|_  _|_ \___
          |___||___||___|
         /  |    |    |  \___ <----- Branches
       ... ...  ...  ... |___| <---- Leaf

    Fig. 1: Tree like organization of DIS
```

Each node contains references to sub-nodes and references to
information it contains. The information can be of any type.
Sub-nodes and information are referenced by an unsigned 64 bit
integer *identifier* assigned incrementally by starting from 1.

```text
             |      Node     |
SubNode <----|* 1         1 *|----> image
SubNode <----|* 2         2 *|----> text
SubNode <----|* 3         3 *|----> music
             |  :         4 -|
             |  :         5 *|----> DIS certificate
             |  :         :  |
             |  :  Local  :  |
             |  Information  |
             |   Reference   |
             |_______________|
   Node references        Information

 Fig. 2: A Node containing information or sub-node references
```

A DIR is a sequence of at most 7 identifiers that defines a
downward path in the DIS node tree. The last ID of a DIR is
the information identifier, all other ID are node IDs. A DIR
without any ID and a length of 0 is a nil DIR.

When the information ID is 0, the DIR is a node reference,
otherwise it is an information reference. When the DIR has
more than 1 ID and the first is 0, the DIR is relative,
otherwise it is an absolute DIR. An absolute DIR is a path
starting at the root. A relative DIR is a path starting at
a node identified by the context.

A human readable string representation of a DIR with ID 1,
2345 and 0 has form "dir:1.2345.0". This example is an
absolute DIR identifying a node.

## Encodings

A DIR is a slice of uint64 but it wouldn't guarantee that it
is valid. It is thus wrapped in a struct that guarantee that
it remains valid.

The binary encoding of a DIR is simply the sequence of binary
encoded identifiers. It is intended to be prefixed by the byte
length but this information is not part of the DIR binary
encoding. Each identifier is encoded in a LEB-128 variant where
the 8 most significant bits are encoded in a single byte. The
maximum byte length of a binary encoded DIR is 9 bytes.

The URI encoding of a DIR starts with "dis:" and ends with "/".
The identifiers separated by a dot are encoded in a LEB-64 encoding
in between. Each group of 6 bits are substituted with the ASCII
character from the string
"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"

The above DIR example "dir:1.2345.0" is encoded as the URI
"dis:1.fa.0/".
