(def l '(1 2 3))
(def b 4)

(assert (== ^(0 ~@l ~b) '(0 1 2 3 4)))

;; note that we use ^ caret to start a template,
;; as opposed to the traditional lisp `` backtick.
;; This lets us use Go-style string literals that
;; are demarcated by backticks.
(defmac when [predicate & body]
  ^(cond ~predicate
      (begin
        ~@body) '()))

(assert (null? (when false 'c)))
(assert (== 'a (when true 'c 'b 'a)))
