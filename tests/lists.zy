(assert (== '(1 2 3) (cons 1 '(2 3))))

(assert (== 1 (first '(1 2 3))))

(assert (== '(2 3) (rest '(1 2 3))))

(assert (== 2 (first (rest '(1 2 3)))))

(let [a 3]
  (assert (== '(0 3) (list 0 a))))

(assert (== '(1 2 4 5) (concat '(1 2) '(4 5))))

; test not-list pairs
(assert (== '(1 \ 2) (cons 1 2)))
(assert (== 2 (rest '(1 \ 2))))
(assert (not (list? '(1 \ 2))))
(assert (list? '()))
(assert (list? '(1 2 3)))
(assert (empty? '()))
(assert (not (empty? '(1 2))))
