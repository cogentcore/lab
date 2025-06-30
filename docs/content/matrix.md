+++
Categories = ["Tensor"]
+++

**Matrix** provides standard 2D linear algebra functions on [[tensor]]s, using [gonum](https://github.com/gonum/gonum) functions for the implementations.

Basic matrix multiplication:

```Goal
##
a := linspace(1., 4., 4., true).reshape(2, 2)
v := [2., 3.]

b  := matrix.Mul(a, a)
c  := matrix.Mul(a, v)
d  := matrix.Mul(v, a)
##

fmt.Println("a:", a)
fmt.Println("b:", b)
fmt.Println("c:", c)
fmt.Println("d:", d)
```

And other standard matrix operations:

```Goal
##
a := linspace(1., 4., 4., true).reshape(2, 2)

t  := tensor.Transpose(a)
d  := matrix.Det(a)
i  := matrix.Inverse(a)
##

fmt.Println("t:", t)
fmt.Println("d:", d)
fmt.Println("i:", i)
```

Including eigenvector functions:

<!--- TODO: not working with 2 return values here: -->

```Goal
##
a := [[2., 1.], [1., 2.]]

v := matrix.EigSym(a)
##

fmt.Println("a:", a)
fmt.Println("v:", v)
```

