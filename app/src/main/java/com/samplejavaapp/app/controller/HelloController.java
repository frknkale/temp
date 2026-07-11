package com.samplejavaapp.app.controller;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
public class HelloController {

    @GetMapping("/**")
    public String handleRequest(@RequestParam Map<String, String> queryParameters) {
        System.out.println("Query Parameters: " + queryParameters);
        return "Query Parameters Logged: " + queryParameters.toString();
    }
}
