(setq org-publish-project-alist
	  '(("org"
		 :base-directory "."
		 :recursive t
		 :publishing-directory ".."
		 :publishing-function org-html-publish-to-html
		 :with-author nil
		 :section-numbers nil
		 :with-toc nil
		 :html-head "<link rel=\"stylesheet\" href=\"bcpdoc.css\" type=\"text/css\"/>")))
