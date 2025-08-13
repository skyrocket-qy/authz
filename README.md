<img src="manifest/icon/f2.png" alt="icon" width="140">

# Zanzibar reversed-engineering

Insipired by [Zanzibar](!https://research.google/pubs/zanzibar-googles-consistent-global-authorization-system/), This project is to implement a distributed authorization read-heavy system

```math
\text{CHECK}(U, \langle \text{object}\#\text{perm} \rangle) = \text{EVALUATE}(\text{perm\_expr}, U, \text{object})
```

Where perm_expr is a boolean expression derived from the schema for perm, composed of relations and logical operators.

Let R₁, R₂, ..., Rₙ be relation names. Then:

```
example perm_expr = Union(R1, Intersection(R2, R3))
```
that is
```math
\text{EVALUATE}(\text{Union}(R_1, \text{Intersection}(R_2, R_3)), U, \text{object}) \\ = 
\text{CHECK}(U, \langle \text{object}\#R_1 \rangle) \lor (\text{CHECK}(U, \langle \text{object}\#R_2 \rangle) \land \text{CHECK}(U, \langle \text{object}\#R_3 \rangle))
```

and each check is

EVALUATE(Union(R1, ..., Rn), U, obj)         = ∨ᵢ CHECK(U, ⟨obj#Ri⟩)

EVALUATE(Intersection(R1, ..., Rn), U, obj)  = ∧ᵢ CHECK(U, ⟨obj#Ri⟩)

EVALUATE(Exclusion(R1, R2), U, obj)          = CHECK(U, ⟨obj#R1⟩) ∧ ¬CHECK(U, ⟨obj#R2⟩)

EVALUATE(Relation(R), U, obj)               = CHECK(U, ⟨obj#R⟩)


CHECK(U, ⟨object#relation⟩) =
∃ tuple ⟨object#relation@U⟩
∨ ∃ tuple ⟨object#relation@U′⟩, where
U′ = ⟨object′#relation′⟩ and CHECK(U, U′)

