import React from "react";
import { blogs } from "../data/blog";
import {Card} from "antd"
function HomePage() {
  return (
    <div style={{alignItems:"center"}}>
      {blogs.map((b) => {
        return (
          <Card
            title={b.title}
            extra={<a href="#">Read Article</a>}
            style={{ width: "87%", margin:"2em 4em" }}
          >
            <p>{b.body}</p>
          </Card>
        );
      })}
    </div>
  );
}

export default HomePage;
