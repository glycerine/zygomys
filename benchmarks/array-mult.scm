(define (mult-vector-loop a b res i)
  (if (= i (vector-length a)) res
    (begin
      (vector-set! res i (* (vector-ref a i) (vector-ref b i)))
      (mult-vector-loop a b res (+ i 1)))))

(define (mult-vector a b)
  (mult-vector-loop a b (make-vector (vector-length a)) 0))

(define (random-vector vec i)
  (if (= i (vector-length vec))
    vec
    (begin
      (vector-set! vec i (random 1.0))
      (random-vector vec (+ i 1)))))

(define (do-in-loop func times)
  (if (= times 0) '()
    (begin
      (func)
      (do-in-loop func (- times 1)))))

(let ((a (random-vector (make-vector 100) 0))
      (b (random-vector (make-vector 100) 0)))
  (do-in-loop (lambda () (mult-vector a b)) 200))

