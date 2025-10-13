import React from "react";
import Container from "./Container";
import Logo from "./Logo";
import HeaderMenu from "./HeaderMenu";
import SearchBar from "./SearchBar";
import CartIcon from "./CartIcon";
import FavoriteButton from "./FavoriteButton";
import SignIn from "./SignIn";
import MobileMenu from "./MobileMenu";

import Link from "next/link";
import { Logs } from "lucide-react";


const Header = async () => {

  let orders = null;


  return (
    <header className="sticky top-0 z-50 py-5 bg-white/70 backdrop-blur-md">
      <Container className="flex items-center justify-between text-lightColor">
        <div className="w-auto md:w-1/3 flex items-center gap-2.5 justify-start md:gap-0">
          <MobileMenu />
          <Logo />
        </div>
        <HeaderMenu />
        <div className="w-auto md:w-1/3 flex items-center justify-end gap-5">
          <SearchBar />
          <CartIcon />
          <FavoriteButton />

        
            <Link
              href={"/orders"}
              className="group relative hover:text-shop_light_green hoverEffect"
            >
              <Logs />
              <span className="absolute -top-1 -right-1 bg-shop_btn_dark_green text-white h-3.5 w-3.5 rounded-full text-xs font-semibold flex items-center justify-center">
           
              </span>
            </Link>
       

        
        </div>
      </Container>
    </header>
  );
};

export default Header;
