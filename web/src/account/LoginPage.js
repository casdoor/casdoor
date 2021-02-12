import React from 'react';
import Face from "../Face";

class LoginPage extends React.Component {
  render() {
    return (
      <Face applicationName={"app-built-in"} account={this.props.account} onLoggedIn={this.props.onLoggedIn.bind(this)} {...this.props} />
    )
  }
}

export default LoginPage;
