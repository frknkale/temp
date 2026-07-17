package com.samplejavaapp.app.controller;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
public class HelloController {

    private static final Logger logger = LoggerFactory.getLogger(HelloController.class);

    @GetMapping("/**")
    public String handleRequest(@RequestParam Map<String, String> queryParameters) {
        logger.info("Query Parameters: {}", queryParameters);
        return "Query Parameters Logged: " + queryParameters.toString();
    }
}
