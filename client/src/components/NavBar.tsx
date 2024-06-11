import { Button } from "@mui/material";
import { useState } from "react";
import PopupForm from "./PopupForm";

export default function NavBar(props: any) {
  const [showPopup, setShowPopup] = useState(0);
  const handleLogout = async () => {
    props.setUser({ name: "", password: "", tasks: [], categories: [] });
    sessionStorage.removeItem("token");
  };
  return (
    <header className="navbar">
      <h1>
        {props.user.name === "" ? "ToDo Planer" : `Hallo ${props.user.name}`}
      </h1>
      <div>
        {props.user.name === "" && (
          <Button variant="contained" onClick={() => setShowPopup(1)}>
            Registrieren
          </Button>
        )}
        {props.user.name === "" && (
          <Button variant="contained" onClick={() => setShowPopup(2)}>
            Login
          </Button>
        )}
        {props.user.name !== "" && (
          <Button variant="contained" onClick={handleLogout}>
            Logout
          </Button>
        )}
      </div>
      <PopupForm
        showPopup={showPopup}
        setShowPopup={setShowPopup}
        setUser={props.setUser}
      />
    </header>
  );
}
