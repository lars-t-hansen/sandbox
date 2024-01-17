;; Workaround for package system on older emacsen, see https://stable.melpa.org/#/getting-started

(if (or (< emacs-major-version 26)
	(and (= emacs-major-version 26) (< emacs-minor-version 3)))
    (progn
      (message "Dropping TLS version b/c old emacs")
      (setq gnutls-algorithm-priority "NORMAL:-VERS-TLS1.3")))

(require 'package)
(add-to-list 'package-archives '("melpa" . "https://melpa.org/packages/"))
(add-to-list 'package-archives
             '("melpa-stable" . "https://stable.melpa.org/packages/") t)
(package-initialize)

(defun tool-bar-off ()
  (if (fboundp 'tool-bar-mode)
      (if (>= emacs-major-version 24)
	  (tool-bar-mode -1)
	(tool-bar-mode nil))))

(defun scroll-bar-off ()
  (if (fboundp 'scroll-bar-mode)
      (if (>= emacs-major-version 24)
	  (scroll-bar-mode -1)
	(scroll-bar-mode nil))))

(defun menu-bar-off ()
  (if (fboundp 'menu-bar-mode)
      (if (>= emacs-major-version 24)
	  (menu-bar-mode -1)
	(menu-bar-mode nil))))

(tool-bar-off)
(scroll-bar-off)
(menu-bar-off)

(put 'upcase-region 'disabled nil)

(push '("\\.cu\\'" . c++-mode) auto-mode-alist)
;(push '("\\.cf\\'" . java-mode) auto-mode-alist)
;(push '("\\.flat_js\\'" . javascript-mode) auto-mode-alist)
;(push '("\\.ts\\'" . javascript-mode) auto-mode-alist)

(defvar c-default-style
  '((c-mode . "stroustrup")
    (c++-mode . "stroustrup")
    (java-mode . "java")))

(defun standard-settings (tabs)
  (set-variable 'fill-column 100)
  (set-variable 'show-trailing-whitespace t)
  (set-variable 'indent-tabs-mode tabs))

(add-hook 'js-mode-hook
	  (lambda ()
	    (standard-settings nil)))

(add-hook 'java-mode-hook
	  (lambda ()
	    (standard-settings nil)
	    (set-variable 'c-basic-offset 4)))

(add-hook 'c-mode-hook
	  (lambda ()
	    (standard-settings nil)))

(add-hook 'c++-mode-hook
	  (lambda ()
	    (standard-settings nil)))

(add-hook 'c-mode-common-hook
          (lambda ()
             (c-set-offset 'case-label '2)
	     (c-set-offset 'statement-case-intro '2)))

(add-hook 'go-mode-hook
	  (lambda ()
	    (standard-settings t)
	    (set-variable 'tab-width 4)))

(add-hook 'rust-mode-hook
	  (lambda ()
	    (standard-settings nil)))

(add-hook 'emacs-lisp-mode-hook
	  (lambda ()
	    (standard-settings nil)))

(add-hook 'markdown-mode-hook
	  (lambda ()
	    (standard-settings nil)))

(add-hook 'sh-mode-hook
	  (lambda ()
	    (standard-settings nil)))

;;(require 'lsp-mode)
;;(setq lsp-enable-snippet nil)

;; Disable vc integration for find-file, it's mostly annoying, especially in large repositories.  If
;; memory serves, this got turned of for the Firefox source tree.
;;
;; Belt and suspenders on this one

(eval-after-load "vc" '(remove-hook 'find-file-hooks 'vc-find-file-hook))
(remove-hook 'find-file-hooks 'vc-find-file-hook)

;; Source grep

;; TODO: would be helpful for files to be sorted by basename first, extension last
;; TODO: should exclude misc benchmarking directories, notably octane (many false hits)
;; TODO: should exclude build directories
;; TODO: probably useful to have a 'cgrep' variant that excludes all js code
;; TODO: a variant 'dgrep' should take an identifier and try to find candidates
;;       for its definition.  This would have to be heuristic, and a number of
;;       the heuristics would be to reject candidates.
;; TODO: should maintain a separate window for each search term (*grep foo*, *grep bar*)
;;       to simplify recursive searches
;; TODO: when the buffer is *grep* we should really not fall back to the default,
;;       but should look in the first line to see if there's a directory there
;;       that matches our criteria.

(defvar *sgrep-default-dir* "/home/lhansen/m-i/js")
(defvar *sgrep-files* "*.h *.c *.cpp *.js")
(defvar *cgrep-files* "*.h *.c *.cpp")

(defun sgrep (pattern)
  "Recursive grep across *sgrep-files* within *sgrep-dir*."
  (interactive 
   (progn
     (grep-compute-defaults)		; A hack - forces grep to be loaded
     (list (let* ((def (current-word))
		  (prompt (if (null def)
			      "Find: "
			    (concat "Find (default " def "): "))))
	     (read-string prompt nil nil def)))))
  (rgrep pattern *sgrep-files* (compute-sgrep-dir (buffer-file-name))))

(defun cgrep (pattern)
  "Recursive grep across *cgrep-files* within *sgrep-dir*."
  (interactive 
   (progn
     (grep-compute-defaults)		; A hack - forces grep to be loaded
     (list (let* ((def (current-word))
		  (prompt (if (null def)
			      "Find: "
			    (concat "Find (default " def "): "))))
	     (read-string prompt nil nil def)))))
  (rgrep pattern *cgrep-files* (compute-sgrep-dir (buffer-file-name))))

(defun compute-sgrep-dir (fn)
  (if (not fn)
      *sgrep-default-dir*
    (let ((dir (file-name-directory fn)))
      (if (not dir)
	  *sgrep-default-dir*
	(setq dir (directory-file-name dir))
	(while (and dir
		    (let ((base (file-name-nondirectory dir)))
		      (and base 
			   (not (string-match-p "^m(ozilla)?-[a-z]+" base)))))
	  (if (not (string-match-p "m(ozilla)?-[a-z]+" dir))
	      (setq dir nil)
	    (let ((ndir (file-name-directory dir)))
	      (setq dir (and ndir (directory-file-name ndir))))))
	(if dir
	    (concat (file-name-as-directory dir) "js")
	  *sgrep-default-dir*)))))
