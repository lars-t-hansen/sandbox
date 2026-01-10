;;; A compiler for basically r5rs scheme, currently with only a limited set of syntax, into a tree
;;; form that is compatible with Twobit's LAP format.

(define (integrate-usual-procedures) #t)

(define (s16? x)
  (and (fixnum? x) (<= -32768 x 37267)))

(define (integrate-relop op args)
  (case (length args)
    ((0) ???)
    ((1) #t)
    ((2) (cons op args))
    (else ...)))

;;; Assoc list from global name of variable-arity or macro-like primitive operation to procedure
;;; that will return a list of only the primitive arities.
(define integrable-procedures
  `((<    1.0 ,integrate-relop)
    (<=   1.0 ,integrate-relop)
    (caar 1   ,(lambda )
    ...))

;;; Assoc list from global name to (arity imm-predicate-or-#f prim-name).
(define primitive-procedures
  `((+       2 ,s16? +)
    (-       2 ,s16? +)
    (*       2 ,s16? +)
    (=       2 ,s16? +)
    (<       2 ,s16? x)
    (<=      2 ,s16? x)
    (>       2 ,s16? x)
    (>=      2 ,s16? x)
    (car     1 #f    car)
    (cdr     1 #f    cdr)
    (fixnum? 1 #f    fixnum?)
    ...))

