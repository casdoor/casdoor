import * as Setting from "./Setting";
import {Redirect, Route, Switch} from "react-router-dom";
import SignupPage from "./auth/SignupPage";
import SelfLoginPage from "./auth/SelfLoginPage";
import LoginPage from "./auth/LoginPage";
import SelfForgetPage from "./auth/SelfForgetPage";
import ForgetPage from "./auth/ForgetPage";
import React from "react";

class EntryPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      application: null,
    };
  }

  renderHomeIfLoggedIn(component) {
    if (this.props.account !== null && this.props.account !== undefined) {
      return <Redirect to="/" />;
    } else {
      return component;
    }
  }

  render() {
    const onUpdateApplication = (application) => {
      this.setState({
        application: application,
      });
    };

    return <div className="loginBackground" style={{backgroundImage: Setting.inIframe() || Setting.isMobile() ? null : `url(${this.state.application?.formBackgroundUrl})`}}>
      <Switch>
        <Route exact path="/signup" render={(props) => this.renderHomeIfLoggedIn(<SignupPage account={this.props.account} onUpdateApplication={onUpdateApplication} {...props} />)} />
        <Route exact path="/signup/:applicationName" render={(props) => this.renderHomeIfLoggedIn(<SignupPage account={this.props.account} onUpdateApplication={onUpdateApplication} {...props} />)} />
        <Route exact path="/login" render={(props) => this.renderHomeIfLoggedIn(<SelfLoginPage account={this.props.account} onUpdateApplication={onUpdateApplication} {...props} />)} />
        <Route exact path="/login/:owner" render={(props) => this.renderHomeIfLoggedIn(<SelfLoginPage account={this.props.account} onUpdateApplication={onUpdateApplication} {...props} />)} />
        <Route exact path="/auto-signup/oauth/authorize" render={(props) => <LoginPage account={this.props.account} type={"code"} mode={"signup"} onUpdateApplication={onUpdateApplication}{...props} />} />
        <Route exact path="/signup/oauth/authorize" render={(props) => <SignupPage account={this.props.account} {...props} onUpdateApplication={onUpdateApplication} />} />
        <Route exact path="/login/oauth/authorize" render={(props) => <LoginPage account={this.props.account} type={"code"} mode={"signin"} onUpdateApplication={onUpdateApplication} {...props} />} />
        <Route exact path="/login/saml/authorize/:owner/:applicationName" render={(props) => <LoginPage account={this.props.account} type={"saml"} mode={"signin"} onUpdateApplication={onUpdateApplication} {...props} />} />
        <Route exact path="/forget" render={(props) => this.renderHomeIfLoggedIn(<SelfForgetPage onUpdateApplication={onUpdateApplication} {...props} />)} />
        <Route exact path="/forget/:applicationName" render={(props) => this.renderHomeIfLoggedIn(<ForgetPage onUpdateApplication={onUpdateApplication} {...props} />)} />
      </Switch>
    </div>;
  }
}

export default EntryPage;
