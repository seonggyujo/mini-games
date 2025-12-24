import { Link } from 'react-router-dom';
import './NavBar.css';

function NavBar() {
  return (
    <nav className="navbar">
      <Link to="/" className="navbar-logo pixel-font">
        MINI GAMES
      </Link>
      <div className="navbar-links">
        <Link to="/" className="nav-link">Home</Link>
        <Link to="/ranking" className="nav-link">Ranking</Link>
      </div>
    </nav>
  );
}

export default NavBar;
