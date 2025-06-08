import logging
from pathlib import Path
from typing import List, Dict, Optional
import git

logger = logging.getLogger(__name__)

class DocsCache:
    """Local cache of the New Relic documentation repository."""

    def __init__(self, cache_dir: Optional[Path] = None):
        self.cache_dir = cache_dir or Path.home() / ".newrelic-mcp" / "docs"
        self.repo_url = "https://github.com/newrelic/docs-website.git"
        self.repo: Optional[git.Repo] = None
        self._ensure_cache()

    def _ensure_cache(self):
        try:
            if not self.cache_dir.exists():
                logger.info(f"Cloning docs repository to {self.cache_dir}")
                self.cache_dir.parent.mkdir(parents=True, exist_ok=True)
                self.repo = git.Repo.clone_from(
                    self.repo_url,
                    self.cache_dir,
                    depth=1,
                    single_branch=True,
                )
            else:
                try:
                    self.repo = git.Repo(self.cache_dir)
                    self.repo.remotes.origin.fetch(depth=1)
                except git.InvalidGitRepositoryError:
                    logger.warning(
                        f"Invalid docs repo at {self.cache_dir}, re-cloning..."
                    )
                    import shutil

                    shutil.rmtree(self.cache_dir)
                    self.repo = git.Repo.clone_from(
                        self.repo_url,
                        self.cache_dir,
                        depth=1,
                        single_branch=True,
                    )
        except Exception as e:
            logger.error(f"Failed to prepare docs cache: {e}")

    def search(self, keyword: str, limit: int = 5) -> List[Dict[str, str]]:
        """Search Markdown docs for a keyword."""
        results: List[Dict[str, str]] = []
        keyword_lower = keyword.lower()
        for md in self.cache_dir.glob("**/*.md"):
            if len(results) >= limit:
                break
            try:
                text = md.read_text(encoding="utf-8", errors="ignore")
                index = text.lower().find(keyword_lower)
                if index != -1:
                    excerpt = text[max(0, index - 40) : index + 40].strip()
                    results.append({
                        "path": str(md.relative_to(self.cache_dir)),
                        "excerpt": excerpt,
                    })
            except Exception as e:
                logger.debug(f"Failed reading {md}: {e}")
        return results

    def get_content(self, rel_path: str) -> str:
        """Return raw Markdown content for a documentation file."""
        doc_path = self.cache_dir / rel_path
        if not doc_path.exists():
            return ""
        try:
            return doc_path.read_text(encoding="utf-8", errors="ignore")
        except Exception as e:
            logger.error(f"Failed to read {doc_path}: {e}")
            return ""
